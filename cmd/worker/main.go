package main

import (
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time" // <-- Import time

	"github.com/darmawguna/tirtaapp.git/config"   // Adjust path
	"github.com/darmawguna/tirtaapp.git/services" // Adjust path
	"github.com/darmawguna/tirtaapp.git/worker"   // <-- Import paket worker
	"github.com/robfig/cron/v3"
	// "github.com/spf13/viper" // Tidak diperlukan lagi di sini
)

func main() {
	log.Println("Starting worker application...")
	config.LoadConfig() // Muat .env

	// Inisialisasi Worker (yang akan mengurus DB, Firebase, Repos)
	workerInstance, err := worker.NewWorker()
	if err != nil {
		log.Fatalf("FATAL: Worker initialization failed: %v", err)
	}

	// [PEMBARUAN] Koneksi ke RabbitMQ dengan Retry
	var queueService services.QueueService
	var connErr error
	maxRetries := 10          // Coba maksimal 10 kali
	retryDelay := 5 * time.Second // Jeda 5 detik antar percobaan

	for i := 0; i < maxRetries; i++ {
		queueService = services.NewQueueService()
		connErr = queueService.Connect()
		if connErr == nil {
			log.Println("Successfully connected and set up RabbitMQ.")
			break // Berhasil, keluar dari loop
		}
		log.Printf("WARN: Failed to connect to RabbitMQ (attempt %d/%d): %v. Retrying in %v...", i+1, maxRetries, connErr, retryDelay)
		time.Sleep(retryDelay)
	}
	// Jika masih error setelah semua retry, baru Fatal
	if connErr != nil {
		log.Fatalf("FATAL: Could not connect to RabbitMQ after %d attempts: %v", maxRetries, connErr)
	}
	// Pastikan defer Close ada SETELAH loop berhasil
	defer queueService.Close()

	// Ambil channel dari service yang sudah siap
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