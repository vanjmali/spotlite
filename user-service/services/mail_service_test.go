package services

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wneessen/go-mail"
)

type fakeMailClient struct {
	messages []*mail.Msg
	err      error
}

func (f *fakeMailClient) DialAndSend(messages ...*mail.Msg) error {
	f.messages = append(f.messages, messages...)
	return f.err
}

func TestMailServiceSendAccountVerificationEmail(t *testing.T) {
	client := &fakeMailClient{}
	cfg := MailConfig{
		VerificationEndpoint: "https://example.com/verify",
		MailFromAddress:      "no-reply@example.com",
	}
	ms := InitMailingService(client, cfg)

	err := ms.SendAccountVerificationEmail("user@example.com", "token-123")
	require.NoError(t, err)
	require.Len(t, client.messages, 1)

	msg := client.messages[0]
	require.Equal(t, []string{"<" + cfg.MailFromAddress + ">"}, msg.GetFromString())
	require.Equal(t, []string{"<user@example.com>"}, msg.GetToString())
	require.Equal(t, []string{"Verify your Spotlite account"}, msg.GetGenHeader(mail.HeaderSubject))

	parts := msg.GetParts()
	require.NotEmpty(t, parts)
	body, err := parts[0].GetContent()
	require.NoError(t, err)
	require.Contains(t, string(body), "https://example.com/verify?token=token-123")
}

func TestMailServiceSendLoginOtp(t *testing.T) {
	client := &fakeMailClient{}
	cfg := MailConfig{MailFromAddress: "no-reply@example.com"}
	ms := InitMailingService(client, cfg)

	err := ms.SendLoginOtp("user@example.com", "654321")
	require.NoError(t, err)
	require.Len(t, client.messages, 1)

	msg := client.messages[0]
	require.Equal(t, []string{"<" + cfg.MailFromAddress + ">"}, msg.GetFromString())
	require.Equal(t, []string{"<user@example.com>"}, msg.GetToString())
	require.Equal(t, []string{"Your Spotlite Login Code"}, msg.GetGenHeader(mail.HeaderSubject))

	parts := msg.GetParts()
	require.NotEmpty(t, parts)
	body, err := parts[0].GetContent()
	require.NoError(t, err)
	require.Contains(t, string(body), "654321")
}
