package services

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/spf13/viper"
	"google.golang.org/api/option"
)

type FirebaseService interface {
	Init() error
	SendNotification(token string, title string, body string) (string, error)
}

type firebaseService struct {
	app *firebase.App
}

func NewFirebaseService() FirebaseService {
	return &firebaseService{}
}

// Init melakukan inisialisasi koneksi ke Firebase menggunakan service account.
func (s *firebaseService) Init() error {
	// Ambil path file service account dari Viper
	path := viper.GetString("FIREBASE_SERVICE_ACCOUNT_PATH")
	if path == "" {
		log.Fatal("FIREBASE_SERVICE_ACCOUNT_PATH environment variable not set.")
	}

	opt := option.WithCredentialsFile(path)
	
	var err error
	s.app, err = firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return err
	}
	
	log.Println("Firebase Admin SDK initialized successfully.")
	return nil
}

// SendNotification mengirimkan satu push notification ke satu perangkat.
func (s *firebaseService) SendNotification(token string, title string, body string) (string, error) {
	ctx := context.Background()
	client, err := s.app.Messaging(ctx)
	if err != nil {
		return "", err
	}

	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Token: token, 
	}

	response, err := client.Send(ctx, message)
	if err != nil {
		return "", err
	}

	return response, nil
}