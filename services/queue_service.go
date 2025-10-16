package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/spf13/viper"
)

// Definisikan nama-nama untuk RabbitMQ
const (
	MainExchange       = "reminders_exchange"
	MainQueue          = "reminders_queue"
	DeadLetterExchange = "reminders_dlx"
	DeadLetterQueue    = "reminders_dlq"
)

// ReminderMessage mendefinisikan payload untuk pesan di RabbitMQ.
type ReminderMessage struct {
	ScheduleType string `json:"schedule_type"`
	ScheduleID   uint   `json:"schedule_id"`
	TimeSlot     int    `json:"time_slot,omitempty"`
}

type QueueService interface {
	Connect() error
	PublishMessage(payload ReminderMessage) error
	Close()
}

type queueService struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
}

func NewQueueService() QueueService {
	return &queueService{}
}

func (s *queueService) Connect() error {
	user := viper.GetString("RABBITMQ_USER")
	pass := viper.GetString("RABBITMQ_PASS")
	host := viper.GetString("RABBITMQ_HOST")
	port := viper.GetString("RABBITMQ_PORT")
	connStr := fmt.Sprintf("amqp://%s:%s@%s:%s/", user, pass, host, port)

	var err error
	s.conn, err = amqp091.Dial(connStr)
	if err != nil { return fmt.Errorf("failed to connect to RabbitMQ: %w", err) }

	s.channel, err = s.conn.Channel()
	if err != nil { return fmt.Errorf("failed to open a channel: %w", err) }

	// --- Deklarasi Arsitektur DLX ---

	// 1. Exchange Utama
	err = s.channel.ExchangeDeclare(MainExchange, "direct", true, false, false, false, nil)
	if err != nil { return fmt.Errorf("failed to declare main exchange: %w", err) }
	
	// 2. Dead Letter Exchange (DLX)
	err = s.channel.ExchangeDeclare(DeadLetterExchange, "direct", true, false, false, false, nil)
	if err != nil { return fmt.Errorf("failed to declare dead letter exchange: %w", err) }

	// 3. Dead Letter Queue (DLQ) - Antrian untuk pesan yang "ditunda"
	_, err = s.channel.QueueDeclare(DeadLetterQueue, true, false, false, false, amqp091.Table{
		"x-message-ttl":             int32(60000), // Pesan di sini akan hidup 1 menit sebelum dicek lagi
		"x-dead-letter-exchange":    MainExchange, // Setelah TTL habis, kirim kembali ke Exchange Utama
	})
	if err != nil { return fmt.Errorf("failed to declare dead letter queue: %w", err) }

	// 4. Binding antara DLX dan DLQ
	err = s.channel.QueueBind(DeadLetterQueue, "", DeadLetterExchange, false, nil)
	if err != nil { return fmt.Errorf("failed to bind dead letter queue: %w", err) }

	// 5. Queue Utama, sekarang dengan argumen DLX
	_, err = s.channel.QueueDeclare(MainQueue, true, false, false, false, amqp091.Table{
		"x-dead-letter-exchange": DeadLetterExchange, // Jika pesan di-reject (Nack) dari sini, kirim ke DLX
	})
	if err != nil { return fmt.Errorf("failed to declare main queue: %w", err) }

	// 6. Binding antara Exchange utama dan Queue utama
	err = s.channel.QueueBind(MainQueue, "", MainExchange, false, nil)
	if err != nil { return fmt.Errorf("failed to bind main queue: %w", err) }

	log.Println("Successfully connected and set up RabbitMQ with DLX architecture")
	return nil
}

// PublishMessage mengirim pesan langsung ke exchange utama.
func (s *queueService) PublishMessage(payload ReminderMessage) error {
	body, err := json.Marshal(payload)
	if err != nil { return fmt.Errorf("failed to marshal payload: %w", err) }

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = s.channel.PublishWithContext(ctx,
		MainExchange, // Kirim ke exchange utama
		"",           // routing key
		false,        // mandatory
		false,        // immediate
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp091.Persistent,
		},
	)
	if err != nil { return fmt.Errorf("failed to publish a message: %w", err) }

	log.Printf("Published message for schedule ID %d\n", payload.ScheduleID)
	return nil
}

func (s *queueService) Close() {
	if s.channel != nil { s.channel.Close() }
	if s.conn != nil { s.conn.Close() }
}