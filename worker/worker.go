package worker

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/darmawguna/tirtaapp.git/config"       // Adjust path if needed
	models "github.com/darmawguna/tirtaapp.git/model" // Adjust path if needed
	"github.com/darmawguna/tirtaapp.git/repositories"
	"github.com/darmawguna/tirtaapp.git/services"
	"github.com/rabbitmq/amqp091-go"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

// --- Struct Worker untuk Mengelola Dependensi ---
type Worker struct {
	db                       *gorm.DB
	firebaseService          services.FirebaseService
	userRepo                 repositories.UserRepository
	deviceRepo               repositories.DeviceRepository
	drugScheduleRepo         repositories.DrugScheduleRepository
	controlScheduleRepo      repositories.ControlScheduleRepository
	hemodialysisScheduleRepo repositories.HemodialysisScheduleRepository
}

// Error khusus untuk memicu requeue via DLX
var ErrRequeueMessage = errors.New("requeue message for later via DLX")

// Constructor untuk Worker
func NewWorker() (*Worker, error) {
	// Koneksi DB & Migrasi
	db := config.ConnectDB()
	config.RunMigration(db,
		&models.User{}, &models.Device{}, &models.DrugSchedule{},
		&models.ControlSchedule{}, &models.HemodialysisSchedule{}, &models.HemodialysisMonitoring{},
	)

	// Inisialisasi Firebase
	firebaseSvc := services.NewFirebaseService()
	if err := firebaseSvc.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase: %w", err)
	}

	// Inisialisasi Repositories
	w := &Worker{
		db:                       db,
		firebaseService:          firebaseSvc,
		userRepo:                 repositories.NewUserRepository(db),
		deviceRepo:               repositories.NewDeviceRepository(db),
		drugScheduleRepo:         repositories.NewDrugScheduleRepository(db),
		controlScheduleRepo:      repositories.NewControlScheduleRepository(db),
		hemodialysisScheduleRepo: repositories.NewHemodialysisScheduleRepository(db),
	}
	log.Println("Worker dependencies initialized.")
	return w, nil
}

// --- Logika Pemroses Pesan RabbitMQ ---
func (w *Worker) MessageHandler(body []byte) error {
	var msg services.ReminderMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		return fmt.Errorf("could not unmarshal message body: %w", err)
	}

	user, location, scheduleDate, err := w.getUserAndTimezone(msg)
	if err != nil {
		log.Printf("Discarding message: %v", err)
		return nil // Return nil to ACK and discard
	}

	switch msg.ScheduleType {
	case "DRUG":
		schedule, err := w.drugScheduleRepo.FindByID(msg.ScheduleID)
		if err != nil { return nil }
		if !schedule.IsActive || isDrugNotificationSent(schedule, msg.TimeSlot) { return nil }

		reminderTime := time.Date(scheduleDate.Year(), scheduleDate.Month(), scheduleDate.Day(), msg.TimeSlot, 0, 0, 0, location)
		notificationTime := reminderTime.Add(-1 * time.Hour)

		if time.Now().Before(notificationTime) { return ErrRequeueMessage }

		devices, _ := w.deviceRepo.FindAllByUserID(user.ID)
		title := "💊 Pengingat Minum Obat"
		body := fmt.Sprintf("Saatnya minum obat %s (dosis: %s) pada pukul %02d:00.", schedule.DrugName, schedule.Dose, msg.TimeSlot)

		w.sendToDevices(devices, title, body)
		if err := w.updateDrugSentStatus(schedule, msg.TimeSlot); err != nil { return err }

	case "KONTROL":
		schedule, err := w.controlScheduleRepo.FindByID(msg.ScheduleID)
		if err != nil { return nil }
		if !schedule.IsActive || schedule.NotificationSent { return nil }

		scheduleTime := time.Date(scheduleDate.Year(), scheduleDate.Month(), scheduleDate.Day(), 7, 0, 0, 0, location)
		notificationTime := scheduleTime.AddDate(0, 0, -1)

		if time.Now().Before(notificationTime) { return ErrRequeueMessage }

		devices, _ := w.deviceRepo.FindAllByUserID(user.ID)
		title := "🗓️ Pengingat Jadwal Kontrol"
		body := fmt.Sprintf("Jangan lupa, Anda memiliki jadwal kontrol besok (%s).", schedule.ControlDate.Format("02 Jan 2006"))

		w.sendToDevices(devices, title, body)
		schedule.NotificationSent = true
		if _, err := w.controlScheduleRepo.Update(schedule); err != nil {
			log.Printf("ERROR: Failed to update sent status for control schedule ID %d: %v", schedule.ID, err)
			return err
		}

	case "HEMODIALISA":
		schedule, err := w.hemodialysisScheduleRepo.FindByID(msg.ScheduleID)
		if err != nil { return nil }
		if !schedule.IsActive || schedule.NotificationSent { return nil }

		scheduleTime := time.Date(scheduleDate.Year(), scheduleDate.Month(), scheduleDate.Day(), 7, 0, 0, 0, location)
		notificationTime := scheduleTime.AddDate(0, 0, -1)

		if time.Now().Before(notificationTime) { return ErrRequeueMessage }

		devices, _ := w.deviceRepo.FindAllByUserID(user.ID)
		title := "🩸 Pengingat Jadwal Hemodialisa"
		body := fmt.Sprintf("Jangan lupa, Anda memiliki jadwal hemodialisa besok (%s).", schedule.ScheduleDate.Format("02 Jan 2006"),)

		w.sendToDevices(devices, title, body)
		schedule.NotificationSent = true
		if _, err := w.hemodialysisScheduleRepo.Update(schedule); err != nil {
			log.Printf("ERROR: Failed to update sent status for hemodialysis schedule ID %d: %v", schedule.ID, err)
			return err
		}
	}
	return nil // Sukses
}

// --- Logika Cron Job ---
func (w *Worker) SendDailyMonitoringReminders() {
	log.Println("Cron Job: Running daily check for Hemodialysis monitoring reminders...")

	serverTimezone := viper.GetString("SERVER_TIMEZONE")
	if serverTimezone == "" { serverTimezone = "Asia/Makassar"; log.Println("WARNING: SERVER_TIMEZONE not set...") }
	location, err := time.LoadLocation(serverTimezone)
	if err != nil { log.Printf("Cron Job ERROR: Invalid SERVER_TIMEZONE..."); location, _ = time.LoadLocation("UTC") }

	now := time.Now().In(location)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)

	schedules, err := w.hemodialysisScheduleRepo.FindSchedulesForDateAndNotNotified(today)
	if err != nil { log.Printf("Cron Job ERROR: checking daily schedules: %v", err); return }
	if len(schedules) == 0 { log.Println("Cron Job: No monitoring reminders to send today."); return }

	log.Printf("Cron Job: Found %d schedules...", len(schedules))
	for _, schedule := range schedules {
		if !schedule.IsActive { continue }

		devices, err := w.deviceRepo.FindAllByUserID(schedule.UserID)
		if err != nil || len(devices) == 0 { continue }

		title := "🩸 Pengingat Pemantauan Hemodialisa"
		body := "Jangan lupa untuk mengisi data pemantauan hemodialisis hari ini, ya."

		log.Printf("Cron Job: Sending monitoring reminder for schedule ID %d...", schedule.ID)
		w.sendToDevices(devices, title, body)

		schedule.MonitoringNotificationSent = true
		if _, err = w.hemodialysisScheduleRepo.Update(schedule); err != nil {
			log.Printf("Cron Job ERROR: updating monitoring status for schedule %d: %v", schedule.ID, err)
		}
	}
	log.Printf("Cron Job: Finished sending %d daily monitoring reminders.", len(schedules))
}

// --- Helper Functions (Methods) ---
func (w *Worker) getUserAndTimezone(msg services.ReminderMessage) (models.User, *time.Location, time.Time, error) {
	var userID uint
	var err error
	var scheduleDate time.Time

	switch msg.ScheduleType {
	case "DRUG":
		schedule, err := w.drugScheduleRepo.FindByID(msg.ScheduleID)
		if err != nil { return models.User{}, nil, time.Time{}, fmt.Errorf("schedule not found") }
		userID = schedule.UserID
		scheduleDate = schedule.ScheduleDate
	case "KONTROL":
		schedule, err := w.controlScheduleRepo.FindByID(msg.ScheduleID)
		if err != nil { return models.User{}, nil, time.Time{}, fmt.Errorf("schedule not found") }
		userID = schedule.UserID
		scheduleDate = schedule.ControlDate
	case "HEMODIALISA":
		schedule, err := w.hemodialysisScheduleRepo.FindByID(msg.ScheduleID)
		if err != nil { return models.User{}, nil, time.Time{}, fmt.Errorf("schedule not found") }
		userID = schedule.UserID
		scheduleDate = schedule.ScheduleDate
	default:
		return models.User{}, nil, time.Time{}, fmt.Errorf("unknown schedule type: %s", msg.ScheduleType)
	}

	user, err := w.userRepo.FindByID(userID)
	if err != nil { return models.User{}, nil, time.Time{}, fmt.Errorf("user not found") }

	location, err := time.LoadLocation(user.Timezone)
	if err != nil { location, _ = time.LoadLocation("UTC") }

	nowInUserLocation := time.Now().In(location)
	todayInUserLocation := time.Date(nowInUserLocation.Year(), nowInUserLocation.Month(), nowInUserLocation.Day(), 0, 0, 0, 0, location)
	scheduleDateInLocation := scheduleDate.In(location)
	scheduleDayStart := time.Date(scheduleDateInLocation.Year(), scheduleDateInLocation.Month(), scheduleDateInLocation.Day(), 0, 0, 0, 0, location)
	if scheduleDayStart.Before(todayInUserLocation) && msg.ScheduleType != "DRUG" {
		return models.User{}, nil, time.Time{}, fmt.Errorf("schedule date is in the past")
	}

	return user, location, scheduleDate, nil
}

func (w *Worker) sendToDevices(devices []models.Device, title, body string) {
	for _, device := range devices {
		responseID, err := w.firebaseService.SendNotification(device.FCMToken, title, body)
		if err != nil {
			log.Printf("-> FAILED to send notification to device %s: %v", device.FCMToken, err)
		} else {
			log.Printf("-> SUCCESS! Notification sent to device %s. Firebase Message ID: %s", device.FCMToken, responseID)
		}
	}
}

func (w *Worker) updateDrugSentStatus(schedule models.DrugSchedule, timeSlot int) error {
	switch timeSlot {
	case 6: schedule.At06Sent = true
	case 12: schedule.At12Sent = true
	case 18: schedule.At18Sent = true
	default: return fmt.Errorf("invalid time slot %d", timeSlot)
	}
	_, err := w.drugScheduleRepo.Update(schedule)
	if err != nil { log.Printf("ERROR: Failed to update drug sent status for ID %d: %v", schedule.ID, err); return err }
	return nil
}

// --- Helper Functions (Biasa) ---
// Fungsi ini tidak butuh akses Worker struct, bisa tetap jadi fungsi biasa
func isDrugNotificationSent(schedule models.DrugSchedule, timeSlot int) bool {
	switch timeSlot {
	case 6: return schedule.At06Sent
	case 12: return schedule.At12Sent
	case 18: return schedule.At18Sent
	default: return true
	}
}

// Fungsi ini tidak butuh akses Worker struct, bisa tetap jadi fungsi biasa
func connectRabbitMQ() (*amqp091.Connection, *amqp091.Channel, error) {
	user := viper.GetString("RABBITMQ_USER")
	pass := viper.GetString("RABBITMQ_PASS")
	host := viper.GetString("RABBITMQ_HOST")
	port := viper.GetString("RABBITMQ_PORT")
	connStr := fmt.Sprintf("amqp://%s:%s@%s:%s/", user, pass, host, port)
	conn, err := amqp091.Dial(connStr)
	if err != nil { return nil, nil, err }
	ch, err := conn.Channel()
	if err != nil { return nil, nil, err }
	return conn, ch, nil
}