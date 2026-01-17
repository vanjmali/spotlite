package entities

import (
	"time"

	"github.com/vanjmali/spotlite/common-lib/account"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TokenType defines the purpose of a stored token.
type TokenType string

const (
	// AccountVerification tokens activate a newly created account.
	AccountVerification TokenType = "ACCOUNT_VERIFICATION"
	// PasswordReset tokens allow changing a forgotten password.
	PasswordReset TokenType = "PASSWORD_RESET"
)

// User models a platform user document stored in MongoDB.
type User struct {
	ID                           primitive.ObjectID    `bson:"_id,omitempty"`
	Username                     string                `bson:"username"`
	FirstName                    string                `bson:"first_name"`
	LastName                     string                `bson:"last_name"`
	Email                        string                `bson:"email"`
	Password                     string                `bson:"password"`
	Role                         account.Role          `bson:"role"`
	PasswordLastChanged          time.Time             `bson:"password_last_changed_at"`
	PasswordExpiresAt            time.Time             `bson:"password_expires_at"`
	AccountStatus                account.AccountStatus `bson:"account_status"`
	CreatedAt                    time.Time             `bson:"created_at"`
	UpdatedAt                    time.Time             `bson:"updated_at"`
	EmailVerification            EmailVerification     `bson:"email_verification"`
	OTPCode                      OTPCode               `bson:"otp_code"`
	LastExpiryNotificationSentAt time.Time             `bson:"last_expiry_notification_sent_at,omitempty"`
}

// EmailVerification keeps the account verification token.
type EmailVerification struct {
	Type  TokenType `bson:"type"`
	Token string    `bson:"token"`
}

// OTPCode stores a hashed OTP and its expiry timestamp.
type OTPCode struct {
	Hash   string    `bson:"hash"`
	Expiry time.Time `bson:"expiry"`
}

// PasswordRecoveryToken represents a password reset token stored in MongoDB.
type PasswordRecoveryToken struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"`
	UserID          primitive.ObjectID `bson:"user_id"`
	Email           string             `bson:"email"`
	TokenHash       string             `bson:"token_hash"`       // bcrypt hash for security
	TokenSearchable string             `bson:"token_searchable"` // SHA256 of plaintext for lookup
	ExpiresAt       time.Time          `bson:"expires_at"`
	UsedAt          *time.Time         `bson:"used_at,omitempty"` // null until used
	CreatedAt       time.Time          `bson:"created_at"`
}
