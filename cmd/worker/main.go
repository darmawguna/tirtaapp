package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/darmawguna/tirtaapp.git/config"   // Adjust path
	"github.com/darmawguna/tirtaapp.git/services" // Adjust path
	"github.com/darmawguna/tirtaapp.git/worker"   // <-- Import paket worker baru
	"github.com/rabbitmq/amqp091-go"
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
)

func main() {
	log.Println("Starting worker application...")
	config.LoadConfig() // Muat .env

	// Inisialisasi Worker (yang akan mengurus DB, Firebase, Repos)
	workerInstance, err := worker.NewWorker()
	if err != nil {
		log.Fatalf("FATAL: Worker initialization failed: %v", err)
	}

	// Koneksi RabbitMQ (tetap di main untuk lifecycle management)
	conn, ch, err := connectRabbitMQ() // Gunakan helper connectRabbitMQ
	if err != nil {
		log.Fatalf("FATAL: Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()
	defer ch.Close()
	ch.Qos(1, 0, false)

	// Setup Cron Job
	cr := cron.New()
	// _, err = cr.AddFunc("* * * * *", workerInstance.SendDailyMonitoringReminders) // Panggil method worker
	_, err = cr.AddFunc("0 6 * * *", workerInstance.SendDailyMonitoringReminders) // Panggil method worker
	if err != nil {
		log.Fatalf("FATAL: Could not add cron job: %v", err)
	}
	cr.Start()
	log.Println("Cron job scheduled.")

	// Mulai Consumer RabbitMQ
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
			// Panggil MessageHandler dari instance worker
			err := workerInstance.MessageHandler(d.Body)
			if err != nil {
				// Gunakan ErrRequeueMessage dari paket worker
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
	// Koneksi RabbitMQ akan ditutup oleh defer
	log.Println("Worker exiting.")
}

// Helper connectRabbitMQ tetap di sini atau bisa dipindah ke package config/utils
func connectRabbitMQ() (*amqp091.Connection, *amqp091.Channel, error) {
	user := viper.GetString("RABBITMQ_USER")
	pass := viper.GetString("RABBITMQ_PASS")
	host := viper.GetString("RABBITMQ_HOST")
	port := viper.GetString("RABBITMQ_PORT")
	connStr := fmt.Sprintf("amqp://%s:%s@%s:%s/", user, pass, host, port)
	conn, err := amqp091.Dial(connStr)
	if err != nil { return nil, nil, fmt.Errorf("failed to dial RabbitMQ: %w", err) }
	ch, err := conn.Channel()
	if err != nil { conn.Close(); return nil, nil, fmt.Errorf("failed to open channel: %w", err) }
	return conn, ch, nil
}