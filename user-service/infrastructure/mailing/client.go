package mailing

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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
	tlsPolicy, err := tlsPolicyFromEnv()
	if err != nil {
		return nil, err
	}

	c, err := mail.NewClient(host,
		mail.WithPort(port),
		mail.WithTLSPolicy(tlsPolicy),
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

func tlsPolicyFromEnv() (mail.TLSPolicy, error) {
	switch strings.ToLower(strings.TrimSpace(utils.GetEnv("SMTP_TLS_POLICY", "none"))) {
	case "none":
		return mail.NoTLS, nil
	case "optional":
		return mail.TLSOpportunistic, nil
	case "mandatory":
		return mail.TLSMandatory, nil
	default:
		return mail.NoTLS, fmt.Errorf("invalid SMTP_TLS_POLICY: use one of none|optional|mandatory")
	}
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
