package main

import (
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	// "time" // Tidak perlu time lagi di sini jika sudah di worker package

	"github.com/darmawguna/tirtaapp.git/config"   // Adjust path
	"github.com/darmawguna/tirtaapp.git/services" // Adjust path
	"github.com/darmawguna/tirtaapp.git/worker"   // <-- Import paket worker
	"github.com/robfig/cron/v3"
	// Masih dibutuhkan jika helper dihapus dan dipindah ke sini
)

func main() {
	log.Println("Starting worker application...")
	config.LoadConfig() // Muat .env

	// Inisialisasi Worker (yang akan mengurus DB, Firebase, Repos)
	workerInstance, err := worker.NewWorker()
	if err != nil {
		log.Fatalf("FATAL: Worker initialization failed: %v", err)
	}

	// [PERBAIKAN] Gunakan QueueService untuk koneksi DAN setup
	queueService := services.NewQueueService()
	if err := queueService.Connect(); err != nil {
		log.Fatalf("FATAL: Failed to connect and setup RabbitMQ: %v", err)
	}
	// Defer Close untuk QueueService akan menutup koneksi & channel di akhir
	defer queueService.Close()

	// [PERBAIKAN] Ambil channel HANYA dari QueueService
	ch := queueService.GetChannel()
	if ch == nil {
		log.Fatalf("FATAL: Failed to get RabbitMQ channel from QueueService")
	}
	// Set QoS menggunakan channel yang didapat dari QueueService
	if err := ch.Qos(1, 0, false); err != nil {
		log.Fatalf("FATAL: Failed to set QoS: %v", err)
	}

	// Setup Cron Job
	cr := cron.New()
	_, err = cr.AddFunc("0 6 * * *", workerInstance.SendDailyMonitoringReminders)
	if err != nil {
		log.Fatalf("FATAL: Could not add cron job: %v", err)
	}
	cr.Start()
	log.Println("Cron job scheduled.")

	// Mulai Consumer RabbitMQ menggunakan channel yang didapat dari QueueService
	msgs, err := ch.Consume(services.MainQueue, "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("FATAL: Failed to register a consumer: %v", err)
	}

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Worker is running. Waiting for messages... Press Ctrl+C to exit.")

	// Loop Pemrosesan Pesan RabbitMQ
	go func() {
		for d := range msgs {
			log.Printf("-> Received RabbitMQ message: %s", d.Body)
			err := workerInstance.MessageHandler(d.Body)
			if err != nil {
				if errors.Is(err, worker.ErrRequeueMessage) {
					log.Println("<- Message not yet due. Re-queuing via DLX.")
					d.Nack(false, false)
				} else {
					log.Printf("ERROR: Processing RabbitMQ message failed: %v. Re-queuing to main queue.", err)
					d.Nack(false, true)
				}
			} else {
				log.Println("<- RabbitMQ message processed successfully.")
				d.Ack(false)
			}
		}
	}()

	<-shutdownChan // Tunggu sinyal shutdown
	log.Println("Shutting down worker gracefully...")
	ctx := cr.Stop()
	<-ctx.Done()
	log.Println("Cron jobs stopped.")
	// Koneksi RabbitMQ akan ditutup oleh defer queueService.Close()
	log.Println("Worker exiting.")
}
