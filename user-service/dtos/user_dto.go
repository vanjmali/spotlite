package dtos

type UserRegistrationDto struct {
	Username  string `json:"username" validate:"required,validusername"`
	FirstName string `json:"firstName" validate:"required,alpha,min=2,max=20"`
	LastName  string `json:"lastName" validate:"required,alpha,min=2,max=20"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,strongpassword"`
}

type UserLoginDto struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,strongpassword"`
}
type VerifyLoginOtpDto struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,len=6"`
}
