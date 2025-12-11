package models

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserRole string

type AccountStatus string

const (
	RoleAdmin  UserRole = "ADMIN"
	RoleMember UserRole = "MEMBER"

	StatusActive   AccountStatus = "ACTIVE"
	StatusInactive AccountStatus = "INACTIVE"
)

type User struct {
	ID                  primitive.ObjectID `bson:"_id, omitempty"`
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
}

func (r *UserRole) SetBSON(raw bson.RawValue) error {
	if raw.Type != bsontype.Type(2) {
		return fmt.Errorf("role field is not a BSON String type")
	}
	*r = UserRole(raw.StringValue())
	return nil
}

func (r *UserRole) GetBSON() (interface{}, error) {
	return string(*r), nil
}

func (s *AccountStatus) SetBSON(raw bson.RawValue) error {
	if raw.Type != bsontype.Type(2) {
		return fmt.Errorf("account status field is not a BSON String type")
	}
	*s = AccountStatus(raw.StringValue())
	return nil
}

func (s *AccountStatus) GetBSON() (interface{}, error) {
	return string(*s), nil
}
