package email

import (
	"fmt"
	"net/smtp"

	"github.com/fvaloy/dbanking/notification/internal/broker"
)

type EmailSender struct {
	smtpHost string
	smtpPort string
	username string
	password string
}

func NewEmailSender(host, port string) *EmailSender {
	return &EmailSender{
		smtpHost: host,
		smtpPort: port,
	}
}

func (e *EmailSender) SendPaymentSucceeded(
	event *broker.PaymentSucceededEvent) error {
	from := "test@example.com"
	to := []string{"user@example.com"}

	msg := fmt.Appendf(nil, "Subject: Payment %s succeeded", event.Reference)

	err := smtp.SendMail(e.smtpHost+":"+e.smtpPort, nil, from, to, msg)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}
	return nil
}
