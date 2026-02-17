package mailing

import (
	"context"
	"strconv"

	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/common-lib/utils"
	mail "github.com/wneessen/go-mail"
)

// MailClient is a thin wrapper around the go-mail client to ease dependency injection.
type MailClient struct {
	C *mail.Client
}

// InitClient builds the mail client.
func InitClient(host string, port int, username string, password string) (*mail.Client, error) {
	c, err := mail.NewClient(host,
		mail.WithPort(port),
		mail.WithTLSPolicy(mail.NoTLS), // Change Mandatory to NoTLS or TLSOptional
		// TODO: configurable TLS
		mail.WithSMTPAuth(mail.SMTPAuthPlainNoEnc),
		mail.WithUsername(username),
		mail.WithPassword(password))
	if err != nil {
		logging.Errorf(context.Background(), "failed to create mail client: %s", err)
		return nil, err
	}
	logging.Infof(context.Background(), "mail client initialized successfully")

	return c, nil
}

// InitClientFromEnv builds the mail client using environment variables with sensible defaults.
func InitClientFromEnv() (*mail.Client, error) {
	host := utils.MustGetEnv("SMTP_HOST")
	portStr := utils.MustGetEnv("SMTP_PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}

	user := utils.MustGetEnv("SMTP_USER")
	pass := utils.MustGetEnv("SMTP_PASS")
	return InitClient(host, port, user, pass)
}
