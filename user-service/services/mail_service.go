package services

import (
	"fmt"
	"log"

	"github.com/wneessen/go-mail"
)

type MailService struct {
	c *mail.Client
}

func InitMailingService(client *mail.Client) *MailService {
	ms := MailService{c: client}
	return &ms
}

func (ms *MailService) sendAccountVerificationEmail(mailto string, token string) {
	m := mail.NewMsg()
	if err := m.From("mail@spotlite.com"); err != nil {
		log.Fatalf("failed to set From address: %s", err)
	}

	if err := m.To(mailto); err != nil {
		log.Fatalf("failed to set To address: %s", err)
	}

	m.Subject("Account verification")
	// TODO: change domain to a environment variable...
	m.SetBodyString(mail.TypeTextHTML,
		fmt.Sprintf("<span>Click <a href='http://localhost:8000/verify?token=%s'>here</a> to verify your account.</span>", token))

	if err := ms.c.DialAndSend(m); err != nil {
		log.Fatalf("failed to send mail: %s", err)
	}
}
