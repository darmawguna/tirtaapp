package services

type EmailSender interface {
	Send(toEmail string, subject string, htmlBody string) error
}
