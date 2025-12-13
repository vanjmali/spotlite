package mailing

import (
	"log"

	mail "github.com/wneessen/go-mail"
)

type MailClient struct {
	C *mail.Client
}

func InitClient(host string, port int, username string, password string) (*mail.Client, error) {
	c, err := mail.NewClient(host,
		mail.WithPort(port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(username),
		mail.WithPassword(password),
		mail.WithTLSPolicy(mail.NoTLS))

	if err != nil {
		log.Fatalf("failed to create mail client: %s", err)
		return nil, err
	}
	log.Println("Mail client initialized successfully.")

	return c, nil
}
