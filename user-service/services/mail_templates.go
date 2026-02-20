package services

import (
	"context"
	"embed"
	"html/template"
	"strconv"
	"strings"
	"time"

	"github.com/vanjmali/spotlite/common-lib/logging"
)

const (
	verificationEmailTemplatePath = "templates/verification_email.html"
	loginOtpEmailTemplatePath     = "templates/login_otp_email.html"
	resetEmailTemplatePath        = "templates/password_reset_email.html"
	commonEmailStylesPath         = "templates/email.css"
)

//go:embed templates/*.html templates/*.css
var mailTemplateFS embed.FS

// EmailTemplateData contains the data needed for email templates.
type EmailTemplateData struct {
	VerificationURL string
	OTP             string
	ResetLink       string
	FooterYear      string
	RepoURL         string
	Styles          template.CSS
}

func defaultEmailTemplateData() EmailTemplateData {
	return EmailTemplateData{
		FooterYear: strconv.Itoa(time.Now().Year()),
		RepoURL:    "https://github.com/vanjmali/spotlite",
	}
}

func renderEmailTemplate(path string, data EmailTemplateData) (string, error) {
	styles, err := mailTemplateFS.ReadFile(commonEmailStylesPath)
	if err != nil {
		logging.Errorf(context.Background(), "failed to read email styles template: %v", err)
		return "", err
	}

	//nolint:gosec // Safe: CSS comes from embedded static file bundled at build time.
	data.Styles = template.CSS(styles)

	tmpl, err := template.ParseFS(mailTemplateFS, path)
	if err != nil {
		logging.Errorf(context.Background(), "failed to parse email template %s: %v", path, err)
		return "", err
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		logging.Errorf(context.Background(), "failed to execute email template %s: %v", path, err)
		return "", err
	}

	return buf.String(), nil
}

// RenderVerificationEmail renders the verification email template with the provided data.
func RenderVerificationEmail(verificationURL string) (string, error) {
	data := defaultEmailTemplateData()
	data.VerificationURL = verificationURL

	return renderEmailTemplate(verificationEmailTemplatePath, data)
}

// RenderLoginOtpEmail renders the login OTP email template with the provided data.
func RenderLoginOtpEmail(otp string) (string, error) {
	data := defaultEmailTemplateData()
	data.OTP = otp

	return renderEmailTemplate(loginOtpEmailTemplatePath, data)
}

// RenderPasswordResetEmail renders the password reset email template with the provided data.
func RenderPasswordResetEmail(resetLink string) (string, error) {
	data := defaultEmailTemplateData()
	data.ResetLink = resetLink

	return renderEmailTemplate(resetEmailTemplatePath, data)
}
