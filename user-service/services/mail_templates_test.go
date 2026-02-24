package services

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRenderLoginOtpEmailRendersStylesAndOtp(t *testing.T) {
	html, err := RenderLoginOtpEmail("123456")
	require.NoError(t, err)

	require.NotContains(t, html, "ZgotmplZ")
	require.Contains(t, html, "<style>")
	require.Contains(t, html, ".otp-code")
	require.Contains(t, html, "123456")
}

func TestRenderVerificationEmailRendersLink(t *testing.T) {
	link := "https://example.com/verify?token=abc"

	html, err := RenderVerificationEmail(link)
	require.NoError(t, err)

	require.NotContains(t, html, "ZgotmplZ")
	require.Contains(t, html, link)
	require.Contains(t, html, "Verify Email Address")
}
