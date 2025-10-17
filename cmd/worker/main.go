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

// [PEMBARUAN] Tambahkan userRepo
var (
	firebaseService          services.FirebaseService
	db                       *gorm.DB
	drugScheduleRepo         repositories.DrugScheduleRepository
	controlScheduleRepo      repositories.ControlScheduleRepository
	hemodialysisScheduleRepo repositories.HemodialysisScheduleRepository
	deviceRepo               repositories.DeviceRepository
	userRepo                 repositories.UserRepository // <-- DITAMBAHKAN
)

var ErrRequeueMessage = errors.New("requeue message for later via DLX")

func main() {
	log.Println("Starting worker...")
	config.LoadConfig()

	firebaseService = services.NewFirebaseService()
	if err := firebaseService.Init(); err != nil {
		log.Fatalf("FATAL: Failed to initialize Firebase: %v", err)
	}

	db = config.ConnectDB()
	config.RunMigration(db,
		&models.User{}, &models.Device{}, &models.DrugSchedule{},
		&models.ControlSchedule{}, &models.HemodialysisSchedule{},
	)

	// [PEMBARUAN] Inisialisasi semua repository yang dibutuhkan
	userRepo = repositories.NewUserRepository(db) // <-- DITAMBAHKAN
	deviceRepo = repositories.NewDeviceRepository(db)
	drugScheduleRepo = repositories.NewDrugScheduleRepository(db)
	controlScheduleRepo = repositories.NewControlScheduleRepository(db)
	hemodialysisScheduleRepo = repositories.NewHemodialysisScheduleRepository(db)

	conn, ch, err := connectRabbitMQ()
	if err != nil {
		log.Fatalf("FATAL: Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()
	defer ch.Close()

	ch.Qos(1, 0, false)

	msgs, err := ch.Consume(services.MainQueue, "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("FATAL: Failed to register a consumer: %v", err)
	}

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Worker is running. Waiting for messages... Press Ctrl+C to exit.")

	go func() {
		for d := range msgs {
			log.Printf("-> Received a message: %s", d.Body)
			err := messageHandler(d.Body)
			if err != nil {
				if errors.Is(err, ErrRequeueMessage) {
					log.Println("<- Message not yet due. Re-queuing via DLX.")
					d.Nack(false, false)
				} else {
					log.Printf("ERROR: Processing message failed: %v. Re-queuing to main queue for retry.", err)
					d.Nack(false, true)
				}
			} else {
				log.Println("<- Message processed successfully.")
				d.Ack(false)
			}
		}
	}()

	<-shutdownChan
	log.Println("Shutting down worker gracefully...")
}

// messageHandler sekarang sadar akan zona waktu
func messageHandler(body []byte) error {
	var msg services.ReminderMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		return fmt.Errorf("could not unmarshal message body: %w", err)
	}

	var user models.User
	var err error

	// Ambil data user berdasarkan tipe jadwal untuk mendapatkan zona waktu
	switch msg.ScheduleType {
	case "DRUG":
		schedule, err := drugScheduleRepo.FindByID(msg.ScheduleID)
		if err != nil {
			return nil
		} // Pesan diabaikan jika jadwal tidak ada
		user, err = userRepo.FindByID(schedule.UserID)
	case "KONTROL":
		schedule, err := controlScheduleRepo.FindByID(msg.ScheduleID)
		if err != nil {
			return nil
		}
		user, err = userRepo.FindByID(schedule.UserID)
	case "HEMODIALISA":
		schedule, err := hemodialysisScheduleRepo.FindByID(msg.ScheduleID)
		if err != nil {
			return nil
		}
		user, err = userRepo.FindByID(schedule.UserID)
	default:
		return fmt.Errorf("unknown schedule type: %s", msg.ScheduleType)
	}

	if err != nil {
		return fmt.Errorf("could not find user for message: %v", err)
	}

	// Muat lokasi/zona waktu spesifik milik user
	location, err := time.LoadLocation(user.Timezone)
	if err != nil {
		log.Printf("WARNING: Invalid timezone '%s' for user %d. Falling back to UTC.", user.Timezone, user.ID)
		location, _ = time.LoadLocation("UTC")
	}

	// Proses berdasarkan tipe jadwal
	switch msg.ScheduleType {
	case "DRUG":
		schedule, _ := drugScheduleRepo.FindByID(msg.ScheduleID)
		if !schedule.IsActive || isDrugNotificationSent(schedule, msg.TimeSlot) {
			return nil
		}

		reminderTime := time.Date(schedule.ScheduleDate.Year(), schedule.ScheduleDate.Month(), schedule.ScheduleDate.Day(), msg.TimeSlot, 0, 0, 0, location)
		notificationTime := reminderTime.Add(-1 * time.Hour)

		if time.Now().Before(notificationTime) {
			return ErrRequeueMessage
		}

		devices, _ := deviceRepo.FindAllByUserID(schedule.UserID)
		title := "💊 Pengingat Minum Obat"
		body := fmt.Sprintf("Saatnya minum obat %s (dosis: %s) pada pukul %02d:00.", schedule.DrugName, schedule.Dose, msg.TimeSlot)

		sendToDevices(devices, title, body)
		updateSentStatus(schedule, msg.TimeSlot, drugScheduleRepo)

	case "KONTROL":
		schedule, _ := controlScheduleRepo.FindByID(msg.ScheduleID)
		if !schedule.IsActive || schedule.NotificationSent {
			return nil
		}

		scheduleTime := time.Date(schedule.ControlDate.Year(), schedule.ControlDate.Month(), schedule.ControlDate.Day(), 7, 0, 0, 0, location)
		notificationTime := scheduleTime.AddDate(0, 0, -1)

		if time.Now().Before(notificationTime) {
			return ErrRequeueMessage
		}

		devices, _ := deviceRepo.FindAllByUserID(schedule.UserID)
		title := "🗓️ Pengingat Jadwal Kontrol"
		body := fmt.Sprintf("Jangan lupa, Anda memiliki jadwal kontrol besok (%s). Catatan: %s", schedule.ControlDate.Format("02 Jan 2006"), schedule.Notes)

		sendToDevices(devices, title, body)
		schedule.NotificationSent = true
		controlScheduleRepo.Update(schedule)

	case "HEMODIALISA":
		schedule, _ := hemodialysisScheduleRepo.FindByID(msg.ScheduleID)
		if !schedule.IsActive || schedule.NotificationSent {
			return nil
		}

		scheduleTime := time.Date(schedule.ScheduleDate.Year(), schedule.ScheduleDate.Month(), schedule.ScheduleDate.Day(), 7, 0, 0, 0, location)
		notificationTime := scheduleTime.AddDate(0, 0, -1)

		if time.Now().Before(notificationTime) {
			return ErrRequeueMessage
		}

		devices, _ := deviceRepo.FindAllByUserID(schedule.UserID)
		title := "🩸 Pengingat Jadwal Hemodialisa"
		body := fmt.Sprintf("Jangan lupa, Anda memiliki jadwal hemodialisa besok (%s). Catatan: %s", schedule.ScheduleDate.Format("02 Jan 2006"), schedule.Notes)

		sendToDevices(devices, title, body)
		schedule.NotificationSent = true
		hemodialysisScheduleRepo.Update(schedule)
	}

	return nil
}

// --- Helper Functions ---
func sendToDevices(devices []models.Device, title, body string) {
	for _, device := range devices {
		responseID, err := firebaseService.SendNotification(device.FCMToken, title, body)
		if err != nil {
			log.Printf("-> FAILED to send notification to device %s: %v", device.FCMToken, err)
		} else {
			log.Printf("-> SUCCESS! Notification sent to device %s. Firebase Message ID: %s", device.FCMToken, responseID)
		}
	}
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
