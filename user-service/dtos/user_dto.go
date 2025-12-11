package dtos

type UserRegistrationDto struct {
	Username  string `json:"username" validate:"required,alphanum,min=4,max=20"`
	FirstName string `json:"firstName" validate:"required,alpha,min=2,max=20"`
	LastName  string `json:"lastName" validate:"required,alpha,min=2,max=20"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=10,max=20,containsany=!@#$%^&*,containsnonalphanum,containsdigit,containslower,containsupper"`
}
