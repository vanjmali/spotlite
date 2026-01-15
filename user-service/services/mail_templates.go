package services

import (
	"html/template"
	"strings"
)

// EmailTemplateData contains the data needed for email templates.
type EmailTemplateData struct {
	VerificationURL string
	OTP             string
	ResetLink       string
}

// VerificationEmailTemplate is the HTML template for account verification emails.
const VerificationEmailTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background-color: #121212;
            margin: 0;
            padding: 0;
        }
        .container {
            max-width: 600px;
            margin: 40px auto;
            background-color: #121212;
            border-radius: 8px;
            overflow: hidden;
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.5);
        }
        .content {
            padding: 40px 30px;
            color: #d1d1d1;
            background-color: #121212;
        }
        .content__heading {
            font-size: 22px;
            font-weight: 600;
            color: #f5f5f5;
            margin: 0 0 15px 0;
        }
        .content__text {
            font-size: 16px;
            line-height: 1.6;
            margin: 15px 0;
            color: #d1d1d1;
        }
        .verification-button {
            display: inline-block;
            background-color: #a855f7;
            color: white;
            padding: 12px 32px;
            border-radius: 500px;
            border: none;
            text-decoration: none;
            font-weight: 600;
            font-size: 16px;
            margin: 30px 0;
            transition: all 200ms ease;
            cursor: pointer;
        }
        .verification-button:hover {
            background-color: #7c3aed;
        }
        .footer {
            background-color: #0a0a0a;
            padding: 20px 30px;
            border-top: 1px solid #333333;
            font-size: 13px;
            color: #a8a8a8;
            text-align: center;
        }
        .footer__text {
            margin: 5px 0;
        }
    </style>
</head>
<body>
    <div class="container">
    
        <div class="content">
            <h2 class="content__heading">Welcome to Spotlite!</h2>
            <p class="content__text">Thank you for signing up. To complete your registration, please verify your email address by clicking the button below:</p>
            <center>
                <a href="{{.VerificationURL}}" class="verification-button">Verify Email Address</a>
            </center>
            <p class="content__text" style="font-size: 14px; color: #a8a8a8;">Or copy and paste this link in your browser:</p>
            <p class="content__text" style="font-size: 13px; color: #a8a8a8; word-break: break-all;">{{.VerificationURL}}</p>
            <p class="content__text">This verification link will expire in 24 hours.</p>
            <p class="content__text">If you did not create this account, please ignore this email.</p>
        </div>
        <div class="footer">
            <p class="footer__text">&copy; 2025 Spotlite. All rights reserved.</p>
            <p class="footer__text">Spotlite | Music Streaming Service</p>
        </div>
    </div>
</body>
</html>`

// LoginOtpEmailTemplate is the HTML template for login OTP emails.
const LoginOtpEmailTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background-color: #121212;
            margin: 0;
            padding: 0;
        }
        .container {
            max-width: 600px;
            margin: 40px auto;
            background-color: #121212;
            border-radius: 8px;
            overflow: hidden;
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.5);
        }
        .content {
            padding: 40px 30px;
            color: #d1d1d1;
            background-color: #121212;
        }
        .content__heading {
            font-size: 22px;
            font-weight: 600;
            color: #f5f5f5;
            margin: 0 0 15px 0;
        }
        .content__text {
            font-size: 16px;
            line-height: 1.6;
            margin: 15px 0;
            color: #d1d1d1;
        }
        .otp-container {
            background-color: #121212;
            padding: 20px;
            border-radius: 8px;
            margin: 30px 0;
            text-align: center;
            border: 1px solid #333333;
        }
        .otp-code {
            font-size: 32px;
            font-weight: 700;
            color: #d8b4fe;
            letter-spacing: 8px;
            margin: 0;
        }
        .otp-warning {
            font-size: 14px;
            color: #a8a8a8;
            margin-top: 10px;
        }
        .footer {
            background-color: #0a0a0a;
            padding: 20px 30px;
            border-top: 1px solid #333333;
            font-size: 13px;
            color: #a8a8a8;
            text-align: center;
        }
        .footer__text {
            margin: 5px 0;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="content">
            <h2 class="content__heading">Your Login Code</h2>
            <p class="content__text">Use this code to complete your login. This code will expire in 5 minutes.</p>
            <div class="otp-container">
                <p class="otp-code">{{.OTP}}</p>
                <p class="otp-warning">Do not share this code with anyone.</p>
            </div>
            <p class="content__text">If you did not request this code, you can safely ignore this email.</p>
        </div>
        <div class="footer">
            <p class="footer__text">&copy; 2025 Spotlite. All rights reserved.</p>
            <p class="footer__text">Spotlite | Music Streaming Service</p>
        </div>
    </div>
</body>
</html>`

// RenderVerificationEmail renders the verification email template with the provided data.
func RenderVerificationEmail(verificationURL string) (string, error) {
	tmpl, err := template.New("verification").Parse(VerificationEmailTemplate)
	if err != nil {
		return "", err
	}

	data := EmailTemplateData{
		VerificationURL: verificationURL,
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// RenderLoginOtpEmail renders the login OTP email template with the provided data.
func RenderLoginOtpEmail(otp string) (string, error) {
	tmpl, err := template.New("otp").Parse(LoginOtpEmailTemplate)
	if err != nil {
		return "", err
	}

	data := EmailTemplateData{
		OTP: otp,
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// PasswordResetEmailTemplate is the HTML template for password reset emails.
const PasswordResetEmailTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background-color: #121212;
            margin: 0;
            padding: 0;
        }
        .container {
            max-width: 600px;
            margin: 40px auto;
            background-color: #121212;
            border-radius: 8px;
            overflow: hidden;
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.5);
        }
        .content {
            padding: 40px 30px;
            color: #d1d1d1;
            background-color: #121212;
        }
        .content__heading {
            font-size: 22px;
            font-weight: 600;
            color: #f5f5f5;
            margin: 0 0 15px 0;
        }
        .content__text {
            font-size: 16px;
            line-height: 1.6;
            margin: 15px 0;
            color: #d1d1d1;
        }
        .reset-button {
            display: inline-block;
            background-color: #a855f7;
            color: white;
            padding: 12px 32px;
            border-radius: 500px;
            border: none;
            text-decoration: none;
            font-weight: 600;
            font-size: 16px;
            margin: 30px 0;
            transition: all 200ms ease;
            cursor: pointer;
        }
        .reset-button:hover {
            background-color: #7c3aed;
        }
        .footer {
            background-color: #0a0a0a;
            padding: 20px 30px;
            border-top: 1px solid #333333;
            font-size: 13px;
            color: #a8a8a8;
            text-align: center;
        }
        .footer__text {
            margin: 5px 0;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="content">
            <h2 class="content__heading">Reset Your Password</h2>
            <p class="content__text">We received a request to reset your password. Click the button below to create a new password:</p>
            <center>
                <a href="{{.ResetLink}}" class="reset-button">Reset Password</a>
            </center>
            <p class="content__text" style="font-size: 14px; color: #a8a8a8;">Or copy and paste this link in your browser:</p>
            <p class="content__text" style="font-size: 13px; color: #a8a8a8; word-break: break-all;">{{.ResetLink}}</p>
            <p class="content__text">This password reset link will expire in 15 minutes.</p>
            <p class="content__text">If you did not request this password reset, you can safely ignore this email.</p>
        </div>
        <div class="footer">
            <p class="footer__text">&copy; 2025 Spotlite. All rights reserved.</p>
            <p class="footer__text">Spotlite | Music Streaming Service</p>
        </div>
    </div>
</body>
</html>`

// RenderPasswordResetEmail renders the password reset email template with the provided data.
func RenderPasswordResetEmail(resetLink string) (string, error) {
	tmpl, err := template.New("passwordReset").Parse(PasswordResetEmailTemplate)
	if err != nil {
		return "", err
	}

	data := EmailTemplateData{
		ResetLink: resetLink,
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
