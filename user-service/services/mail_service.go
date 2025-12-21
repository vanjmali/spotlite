package services

import (
	"log"

	"github.com/wneessen/go-mail"
)

// MailService sends transactional emails such as account verification and OTPs.
type MailService struct {
	c *mail.Client
}

// InitMailingService wraps the mail client with a service layer.
func InitMailingService(client *mail.Client) *MailService {
	ms := MailService{c: client}
	return &ms
}

func (ms *MailService) sendAccountVerificationEmail(mailto string, token string) error {
	m := mail.NewMsg()
	if err := m.From("mail@spotlite.com"); err != nil {
		log.Printf("failed to set From address: %v", err)
		return err
	}

	if err := m.To(mailto); err != nil {
		log.Printf("failed to set To address: %v", err)
		return err
	}

	m.Subject("Verify your Spotlite account")

	// Render email template with verification URL
	verificationURL := "http://localhost:3000/api/users/verify?token=" + token
	emailBody, err := RenderVerificationEmail(verificationURL)
	if err != nil {
		log.Printf("failed to render verification email template: %v", err)
		return err
	}

	m.SetBodyString(mail.TypeTextHTML, emailBody)

	err = ms.c.DialAndSend(m)
	if err != nil {
		log.Printf("failed to send verification email: %v", err)
	}
	return err
}

// SendLoginOtp dispatches a one-time password email to the given recipient.
func (ms *MailService) SendLoginOtp(mailto string, otp string) error {
	m := mail.NewMsg()

	if err := m.From("mail@spotlite.com"); err != nil {
		log.Printf("failed to set From address: %v", err)
		return err
	}
	if err := m.To(mailto); err != nil {
		log.Printf("failed to set To address: %v", err)
		return err
	}

	m.Subject("Your Spotlite Login Code")

	// Render email template with OTP
	emailBody, err := RenderLoginOtpEmail(otp)
	if err != nil {
		log.Printf("failed to render OTP email template: %v", err)
		return err
	}

	m.SetBodyString(mail.TypeTextHTML, emailBody)

	err = ms.c.DialAndSend(m)
	if err != nil {
		log.Printf("failed to send OTP email: %v", err)
	}
	return err
}
