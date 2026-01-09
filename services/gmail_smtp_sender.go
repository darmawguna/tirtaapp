package services

import (
	"fmt"
	"strconv"

	"github.com/spf13/viper"
	"gopkg.in/gomail.v2"
)

type GmailSMTPSender struct {
	host      string
	port      int
	username  string
	password  string
	fromName  string
	fromEmail string
}

func NewGmailSMTPSenderFromConfig() (*GmailSMTPSender, error) {
	portStr := viper.GetString("SMTP_PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid SMTP_PORT: %q (%w)", portStr, err)
	}

	s := &GmailSMTPSender{
		host:      viper.GetString("SMTP_HOST"),
		port:      port,
		username:  viper.GetString("SMTP_USER"),
		password:  viper.GetString("SMTP_PASS"),
		fromName:  viper.GetString("SMTP_FROM_NAME"),
		fromEmail: viper.GetString("SMTP_FROM_EMAIL"),
	}

	if s.host == "" || s.username == "" || s.password == "" || s.fromEmail == "" {
		return nil, fmt.Errorf("missing SMTP config (HOST/PORT/USER/PASS/FROM_EMAIL)")
	}

	return s, nil
}

func (s *GmailSMTPSender) Send(toEmail, subject, htmlBody string) error {
	m := gomail.NewMessage()
	if s.fromName != "" {
		m.SetAddressHeader("From", s.fromEmail, s.fromName)
	} else {
		m.SetHeader("From", s.fromEmail)
	}
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlBody)

	d := gomail.NewDialer(s.host, s.port, s.username, s.password)
	return d.DialAndSend(m)
}
