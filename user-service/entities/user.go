package entities

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserRole enumerates the access level of a user.
type UserRole string

// AccountStatus captures whether a user is allowed to interact with the platform.
type AccountStatus string

// TokenType defines the purpose of a stored token.
type TokenType string

const (
	// RoleAdmin represents elevated privileges.
	RoleAdmin UserRole = "ADMIN"
	// RoleMember is the standard user role.
	RoleMember UserRole = "MEMBER"

	// StatusActive means the account is usable.
	StatusActive AccountStatus = "ACTIVE"
	// StatusInactive blocks the account until activation.
	StatusInactive AccountStatus = "INACTIVE"

	// AccountVerification tokens activate a newly created account.
	AccountVerification TokenType = "ACCOUNT_VERIFICATION"
	// PasswordReset tokens allow changing a forgotten password.
	PasswordReset TokenType = "PASSWORD_RESET"
)

// User models a platform user document stored in MongoDB.
type User struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty"`
	Username            string             `bson:"username"`
	FirstName           string             `bson:"first_name"`
	LastName            string             `bson:"last_name"`
	Email               string             `bson:"email"`
	Password            string             `bson:"password"`
	Role                UserRole           `bson:"role"`
	PasswordLastChanged time.Time          `bson:"password_last_changed_at"`
	PasswordExpiresAt   time.Time          `bson:"password_expires_at"`
	AccountStatus       AccountStatus      `bson:"account_status"`
	CreatedAt           time.Time          `bson:"created_at"`
	UpdatedAt           time.Time          `bson:"updated_at"`
	EmailVerification   EmailVerification  `bson:"email_verification"`
	OTPCode             OTPCode            `bson:"otp_code"`
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

// SetBSON implements custom BSON decoding for UserRole.
func (r *UserRole) SetBSON(raw bson.RawValue) error {
	if raw.Type != bsontype.Type(2) {
		return errors.New("role field is not a BSON String type")
	}
	*r = UserRole(raw.StringValue())
	return nil
}

// GetBSON implements custom BSON encoding for UserRole.
func (r *UserRole) GetBSON() (interface{}, error) {
	return string(*r), nil
}

// SetBSON implements custom BSON decoding for AccountStatus.
func (s *AccountStatus) SetBSON(raw bson.RawValue) error {
	if raw.Type != bsontype.Type(2) {
		return errors.New("account status field is not a BSON String type")
	}
	*s = AccountStatus(raw.StringValue())
	return nil
}

// GetBSON implements custom BSON encoding for AccountStatus.
func (s *AccountStatus) GetBSON() (interface{}, error) {
	return string(*s), nil
}

// SetBSON implements custom BSON decoding for TokenType.
func (t *TokenType) SetBSON(raw bson.RawValue) error {
	if raw.Type != bsontype.Type(2) {
		return errors.New("token type field is not a BSON String type")
	}
	*t = TokenType(raw.StringValue())
	return nil
}

// GetBSON implements custom BSON encoding for TokenType.
func (t *TokenType) GetBSON() (interface{}, error) {
	return string(*t), nil
}
