package services

import (
	"fmt"
	"log"
	"net/url"

	"github.com/wneessen/go-mail"
)

// MailConfig holds configuration needed by the mail service.
type MailConfig struct {
	VerificationEndpoint string
	PasswordResetURL     string
	MailFromAddress      string
}

// MailSender defines the email operations required by UserService.
type MailSender interface {
	SendAccountVerificationEmail(mailto string, token string) error
	SendLoginOtp(mailto string, otp string) error
	SendPasswordResetEmail(mailto string, token string) error
}

// MailClient defines the mail client operations used by MailService.
type MailClient interface {
	DialAndSend(messages ...*mail.Msg) error
}

// MailService sends transactional emails such as account verification and OTPs.
type MailService struct {
	c      MailClient
	config MailConfig
}

// InitMailingService wraps the mail client with a service layer.
func InitMailingService(client MailClient, cfg MailConfig) *MailService {
	ms := MailService{c: client, config: cfg}
	return &ms
}

func setFromToAddress(m *mail.Msg, from, to string) error {
	if err := m.From(from); err != nil {
		return fmt.Errorf("failed to set From address: %v", err)
	}

	if err := m.To(to); err != nil {
		return fmt.Errorf("failed to set To address: %v", err)
	}

	return nil
}

func (ms *MailService) SendAccountVerificationEmail(mailto string, token string) error {
	m := mail.NewMsg()
	if err := setFromToAddress(m, ms.config.MailFromAddress, mailto); err != nil {
		return err
	}

	// Render email template with verification URL
	verificationURL := ms.config.VerificationEndpoint + "?token=" + url.QueryEscape(token)
	emailBody, err := RenderVerificationEmail(verificationURL)
	if err != nil {
		log.Printf("failed to render verification email template: %v", err)
		return err
	}

	m.Subject("Verify your Spotlite account")
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
	if err := setFromToAddress(m, ms.config.MailFromAddress, mailto); err != nil {
		return err
	}

	// Render email template with OTP
	emailBody, err := RenderLoginOtpEmail(otp)
	if err != nil {
		log.Printf("failed to render OTP email template: %v", err)
		return err
	}

	m.Subject("Your Spotlite Login Code")
	m.SetBodyString(mail.TypeTextHTML, emailBody)

	err = ms.c.DialAndSend(m)
	if err != nil {
		log.Printf("failed to send OTP email: %v", err)
	}
	return err
}

// SendPasswordResetEmail sends a password reset magic link email to the given recipient.
func (ms *MailService) SendPasswordResetEmail(mailto string, token string) error {
	m := mail.NewMsg()
	if err := setFromToAddress(m, ms.config.MailFromAddress, mailto); err != nil {
		return err
	}

	// Render email template with reset link
	resetLink := ms.config.PasswordResetURL + "?token=" + url.QueryEscape(token)
	emailBody, err := RenderPasswordResetEmail(resetLink)
	if err != nil {
		return fmt.Errorf("failed to render password reset email template: %v", err)
	}

	m.Subject("Reset Your Spotlite Password")
	m.SetBodyString(mail.TypeTextHTML, emailBody)
	err = ms.c.DialAndSend(m)
	return err
}

func (ms *MailService) SendExpiryMail(mailto string) error {
	log.Printf("sending email")
	m := mail.NewMsg()

	if err := m.From("mail@spotlite.com"); err != nil {
		return err
	}
	if err := m.To(mailto); err != nil {
		return err
	}

	m.Subject("Expiry")
	m.SetBodyString(mail.TypeTextHTML, "your password is expiring soon")

	return ms.c.DialAndSend(m)
}
