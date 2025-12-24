package dtos

// UserRegistrationDto represents the payload required to create a new user account.
type UserRegistrationDto struct {
	Username  string `json:"username" validate:"required,validusername"`
	FirstName string `json:"firstName" validate:"required,validname,min=2,max=20"`
	LastName  string `json:"lastName" validate:"required,validname,min=2,max=20"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,strongpassword"`
}

// UserLoginDto holds the credentials submitted when a user signs in.
type UserLoginDto struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,strongpassword"`
}

// VerifyLoginOtpDto carries the email and code for OTP verification during login.
type VerifyLoginOtpDto struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,len=6"`
}

// ResendOtpDto carries the email for resending OTP during login.
type ResendOtpDto struct {
	Email string `json:"email" validate:"required,email"`
}

// CheckEmailDto carries the email to check if it's already registered.
type CheckEmailDto struct {
	Email string `json:"email" validate:"required,email"`
}
