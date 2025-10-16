package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/darmawguna/tirtaapp.git/config"
	models "github.com/darmawguna/tirtaapp.git/model"

	"github.com/darmawguna/tirtaapp.git/repositories"
	"github.com/darmawguna/tirtaapp.git/services"
	"github.com/rabbitmq/amqp091-go"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

// Global context
var (
	firebaseService  services.FirebaseService
	db               *gorm.DB
	drugScheduleRepo repositories.DrugScheduleRepository
	deviceRepo       repositories.DeviceRepository
)

// [BARU] Definisikan error khusus untuk memicu requeue via DLX
var ErrRequeueMessage = errors.New("requeue message for later via DLX")

func main() {
	// --- Tahap 1: Inisialisasi ---
	log.Println("Starting worker...")
	config.LoadConfig()

	firebaseService = services.NewFirebaseService()
	if err := firebaseService.Init(); err != nil {
		log.Fatalf("FATAL: Failed to initialize Firebase: %v", err)
	}

	db = config.ConnectDB()
	drugScheduleRepo = repositories.NewDrugScheduleRepository(db)
	deviceRepo = repositories.NewDeviceRepository(db)

	conn, ch, err := connectRabbitMQ()
	if err != nil {
		log.Fatalf("FATAL: Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()
	defer ch.Close()

	// [BARU] Set QoS (Quality of Service)
	// Ini memastikan worker hanya mengambil 1 pesan dalam satu waktu, mencegah overload.
	err = ch.Qos(1, 0, false)
	if err != nil {
		log.Fatalf("FATAL: Failed to set QoS: %v", err)
	}

	// --- Tahap 2: Mulai Consumer ---
	msgs, err := ch.Consume(
		services.MainQueue, "", false, false, false, false, nil,
	)
	if err != nil {
		log.Fatalf("FATAL: Failed to register a consumer: %v", err)
	}

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Worker is running. Waiting for messages... Press Ctrl+C to exit.")

	// --- Tahap 3: Loop Pemrosesan Pesan ---
	go func() {
		for d := range msgs {
			log.Printf("-> Received a message: %s", d.Body)

			err := messageHandler(d.Body)

			// [PEMBARUAN] Logika Acknowledge yang lebih cerdas
			if err != nil {
				// Cek apakah errornya adalah error khusus untuk requeue via DLX
				if errors.Is(err, ErrRequeueMessage) {
					log.Println("<- Message not yet due. Re-queuing via DLX.")
					// Nack TANPA requeue. Pesan akan otomatis dikirim RabbitMQ ke DLX.
					d.Nack(false, false)
				} else {
					log.Printf("ERROR: Processing message failed: %v. Re-queuing to main queue for retry.", err)
					// Untuk error lain (misal: DB down sementara), requeue ke antrian utama.
					d.Nack(false, true)
				}
			} else {
				log.Println("<- Message processed successfully.")
				// Beri tahu RabbitMQ bahwa pesan sudah selesai diproses.
				d.Ack(false)
			}
		}
	}()

	<-shutdownChan
	log.Println("Shutting down worker gracefully...")
}

// messageHandler adalah jantung dari worker, berisi logika untuk setiap pesan
func messageHandler(body []byte) error {
	var msg services.ReminderMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		return fmt.Errorf("could not unmarshal message body: %w", err)
	}

	switch msg.ScheduleType {
	case "DRUG":
		schedule, err := drugScheduleRepo.FindByID(msg.ScheduleID)
		if err != nil {
			log.Printf("WARNING: Drug schedule ID %d not found in DB, discarding message.", msg.ScheduleID)
			return nil
		}

		if !schedule.IsActive || isDrugNotificationSent(schedule, msg.TimeSlot) {
			log.Printf("Skipping drug schedule ID %d (inactive or already sent).", schedule.ID)
			return nil
		}

		// [PEMBARUAN] Logika Pengecekan Waktu
		// Asumsikan WITA. Ganti "Asia/Makassar" jika perlu.
		location, _ := time.LoadLocation("Asia/Makassar")
		reminderTime := time.Date(schedule.ScheduleDate.Year(), schedule.ScheduleDate.Month(), schedule.ScheduleDate.Day(), msg.TimeSlot, 0, 0, 0, location)
		notificationTime := reminderTime.Add(-1 * time.Hour)

		// Cek apakah waktu sekarang sudah melewati waktu notifikasi yang dijadwalkan
		if time.Now().Before(notificationTime) {
			// Jika belum waktunya, kembalikan error khusus untuk memicu requeue via DLX
			return ErrRequeueMessage
		}

		// --- Jika sudah waktunya, lanjutkan proses pengiriman ---

		devices, err := deviceRepo.FindAllByUserID(schedule.UserID)
		if err != nil || len(devices) == 0 {
			return fmt.Errorf("could not find devices for user ID %d", schedule.UserID)
		}

		title := "💊 Pengingat Minum Obat"
		body := fmt.Sprintf("Saatnya minum obat %s (dosis: %s) pada pukul %02d:00.", schedule.DrugName, schedule.Dose, msg.TimeSlot)

		for _, device := range devices {
			_, err := firebaseService.SendNotification(device.FCMToken, title, body)
			if err != nil {
				log.Printf("-> Failed to send notification to device %s: %v", device.FCMToken, err)
			} else {
				log.Printf("-> Notification sent to device %s", device.FCMToken)
			}
		}

		updateSentStatus(schedule, msg.TimeSlot, drugScheduleRepo)

	case "KONTROL":
		log.Println("Processing 'KONTROL' schedule type... (Not Implemented)")
	case "HEMODIALISA":
		log.Println("Processing 'HEMODIALISA' schedule type... (Not Implemented)")
	}

	return nil
}

// --- Helper Functions (Tidak ada perubahan) ---

func isDrugNotificationSent(schedule models.DrugSchedule, timeSlot int) bool {
	switch timeSlot {
	case 6:
		return schedule.At06Sent
	case 12:
		return schedule.At12Sent
	case 18:
		return schedule.At18Sent
	default:
		return true
	}
}

func updateSentStatus(schedule models.DrugSchedule, timeSlot int, drugRepo repositories.DrugScheduleRepository) {
	switch timeSlot {
	case 6:
		schedule.At06Sent = true
	case 12:
		schedule.At12Sent = true
	case 18:
		schedule.At18Sent = true
	}
	_, err := drugRepo.Update(schedule)
	if err != nil {
		log.Printf("ERROR: Failed to update sent status for drug schedule ID %d: %v", schedule.ID, err)
	}
}

func connectRabbitMQ() (*amqp091.Connection, *amqp091.Channel, error) {
	user := viper.GetString("RABBITMQ_USER")
	pass := viper.GetString("RABBITMQ_PASS")
	host := viper.GetString("RABBITMQ_HOST")
	port := viper.GetString("RABBITMQ_PORT")
	connStr := fmt.Sprintf("amqp://%s:%s@%s:%s/", user, pass, host, port)

	conn, err := amqp091.Dial(connStr)
	if err != nil {
		return nil, nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, nil, err
	}

	return conn, ch, nil
}
