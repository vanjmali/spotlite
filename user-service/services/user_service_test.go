package services

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"github.com/vanjmali/spotlite/common-lib/account"
	"github.com/vanjmali/spotlite/common-lib/clock"
	"github.com/vanjmali/spotlite/user-service/dtos"
	"github.com/vanjmali/spotlite/user-service/entities"
	"github.com/vanjmali/spotlite/user-service/utils/auth"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// Helper function to add user ID to context for testing
func contextWithUserID(ctx context.Context, userID primitive.ObjectID) context.Context {
	return context.WithValue(ctx, ctxKey("userId"), userID.Hex())
}

type ctxKey string

type fakeUserRepo struct {
	existsByUsernameFn func(context.Context, string) (bool, error)
	existsByEmailFn    func(context.Context, string) (bool, error)
	createFn           func(context.Context, entities.User) error
	findUserByEmailFn  func(context.Context, string) (*entities.User, error)
	findUserByIDFn     func(context.Context, primitive.ObjectID) (*entities.User, error)
	setLoginOtpFn      func(context.Context, primitive.ObjectID, string, time.Time) error
	clearLoginOtpFn    func(context.Context, primitive.ObjectID) error
	activeAndRevokeFn  func(context.Context, string) error
	setHashPasswordFn  func(context.Context, primitive.ObjectID, string, time.Time, time.Time) error

	createCalled     bool
	createdUser      entities.User
	setLoginCalled   bool
	setLoginHash     string
	setLoginExpiry   time.Time
	clearLoginCalled bool
	clearLoginUserID primitive.ObjectID
}

func (f *fakeUserRepo) Create(ctx context.Context, user entities.User) error {
	f.createCalled = true
	f.createdUser = user
	if f.createFn != nil {
		return f.createFn(ctx, user)
	}
	return nil
}

func (f *fakeUserRepo) ActiveAndRevokeToken(ctx context.Context, token string) error {
	if f.activeAndRevokeFn != nil {
		return f.activeAndRevokeFn(ctx, token)
	}
	return nil
}

func (f *fakeUserRepo) SetHashPassowrd(ctx context.Context, userId primitive.ObjectID, passwordHash string, newTime, expiresAt time.Time) error {
	if f.setHashPasswordFn != nil {
		return f.setHashPasswordFn(ctx, userId, passwordHash, newTime, expiresAt)
	}
	return nil
}

func (f *fakeUserRepo) SetLoginOtp(ctx context.Context, userId primitive.ObjectID, hash string, expiry time.Time) error {
	f.setLoginCalled = true
	f.setLoginHash = hash
	f.setLoginExpiry = expiry
	if f.setLoginOtpFn != nil {
		return f.setLoginOtpFn(ctx, userId, hash, expiry)
	}
	return nil
}

func (f *fakeUserRepo) ClearLoginOtp(ctx context.Context, userId primitive.ObjectID) error {
	f.clearLoginCalled = true
	f.clearLoginUserID = userId
	if f.clearLoginOtpFn != nil {
		return f.clearLoginOtpFn(ctx, userId)
	}
	return nil
}

func (f *fakeUserRepo) FindUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	if f.findUserByEmailFn != nil {
		return f.findUserByEmailFn(ctx, email)
	}
	return &entities.User{}, nil
}

func (f *fakeUserRepo) FindUsersForExpiryNotification(
	ctx context.Context,
	daysUntilExpiry int,
	batchSize int,
	lastID string,
) ([]*entities.User, string, error) {
	return nil, "", nil
}

func (f *fakeUserRepo) UpdateExpiryNotificationSentDate(ctx context.Context, userID primitive.ObjectID) error {
	return nil
}

func (f *fakeUserRepo) FindUserByID(ctx context.Context, id primitive.ObjectID) (*entities.User, error) {
	if f.findUserByIDFn != nil {
		return f.findUserByIDFn(ctx, id)
	}
	return &entities.User{}, nil
}

func (f *fakeUserRepo) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	if f.existsByUsernameFn != nil {
		return f.existsByUsernameFn(ctx, username)
	}
	return false, nil
}

func (f *fakeUserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	if f.existsByEmailFn != nil {
		return f.existsByEmailFn(ctx, email)
	}
	return false, nil
}

type fakeMailService struct {
	sendVerificationFn func(string, string) error
	sendLoginOtpFn     func(string, string) error

	verificationCalled bool
	verificationMailTo string
	verificationToken  string

	loginOtpCalled bool
	loginOtpMailTo string
	loginOtpCode   string
}

func (f *fakeMailService) SendAccountVerificationEmail(mailto string, token string) error {
	f.verificationCalled = true
	f.verificationMailTo = mailto
	f.verificationToken = token
	if f.sendVerificationFn != nil {
		return f.sendVerificationFn(mailto, token)
	}
	return nil
}

func (f *fakeMailService) SendLoginOtp(mailto string, otp string) error {
	f.loginOtpCalled = true
	f.loginOtpMailTo = mailto
	f.loginOtpCode = otp
	if f.sendLoginOtpFn != nil {
		return f.sendLoginOtpFn(mailto, otp)
	}
	return nil
}

func TestUserServiceRegisterUsernameTaken(t *testing.T) {
	repo := &fakeUserRepo{
		existsByUsernameFn: func(ctx context.Context, username string) (bool, error) {
			return true, nil
		},
	}
	mail := &fakeMailService{}
	svc := NewUserService(repo, mail)

	err := svc.Register(context.Background(), &dtos.UserRegistrationDto{
		Username:  "taken",
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     "jane@example.com",
		Password:  "StrongPass123!",
	})

	require.ErrorIs(t, err, ErrUsernameTaken)
	require.False(t, mail.verificationCalled, "verification email should not be sent")
	require.False(t, repo.createCalled, "user should not be created")
}

func TestUserServiceRegisterEmailTaken(t *testing.T) {
	repo := &fakeUserRepo{
		existsByUsernameFn: func(ctx context.Context, username string) (bool, error) {
			return false, nil
		},
		existsByEmailFn: func(ctx context.Context, email string) (bool, error) {
			return true, nil
		},
	}
	mail := &fakeMailService{}
	svc := NewUserService(repo, mail)

	err := svc.Register(context.Background(), &dtos.UserRegistrationDto{
		Username:  "unique",
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     "taken@example.com",
		Password:  "StrongPass123!",
	})

	require.ErrorIs(t, err, ErrEmailTaken)
	require.False(t, mail.verificationCalled, "verification email should not be sent")
	require.False(t, repo.createCalled, "user should not be created")
}

func TestUserServiceRegisterEmailSendFails(t *testing.T) {
	repo := &fakeUserRepo{}
	mail := &fakeMailService{
		sendVerificationFn: func(string, string) error {
			return errors.New("smtp down")
		},
	}
	svc := NewUserService(repo, mail)

	err := svc.Register(context.Background(), &dtos.UserRegistrationDto{
		Username:  "unique",
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     "jane@example.com",
		Password:  "StrongPass123!",
	})

	require.Error(t, err, "expected error when email sending fails")
	require.True(t, mail.verificationCalled, "verification email should be attempted")
	require.False(t, repo.createCalled, "user should not be created on email failure")
}

func TestUserServiceRegisterSuccess(t *testing.T) {
	repo := &fakeUserRepo{}
	mail := &fakeMailService{}
	svc := NewUserService(repo, mail)

	dto := &dtos.UserRegistrationDto{
		Username:  "unique",
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     "jane@example.com",
		Password:  "StrongPass123!",
	}

	err := svc.Register(context.Background(), dto)
	require.NoError(t, err)
	require.True(t, repo.createCalled, "user should be created")
	require.Equal(t, dto.Email, repo.createdUser.Email)
	require.Equal(t, dto.Username, repo.createdUser.Username)
	require.NotEqual(t, dto.Password, repo.createdUser.Password)
	require.Equal(t, account.StatusInactive, repo.createdUser.AccountStatus)
	require.NotEmpty(t, repo.createdUser.EmailVerification.Token)
}

func TestUserServiceLoginInactive(t *testing.T) {
	repo := &fakeUserRepo{
		findUserByEmailFn: func(ctx context.Context, email string) (*entities.User, error) {
			return &entities.User{AccountStatus: account.StatusInactive}, nil
		},
	}
	svc := NewUserService(repo, &fakeMailService{})

	err := svc.Login(context.Background(), &dtos.UserLoginDto{
		Email:    "user@example.com",
		Password: "StrongPass123!",
	})

	require.ErrorIs(t, err, ErrUserInactive)
}

func TestUserServiceLoginExpiredPassword(t *testing.T) {
	hashed, _ := auth.HashPassword("StrongPass123!")
	repo := &fakeUserRepo{
		findUserByEmailFn: func(ctx context.Context, email string) (*entities.User, error) {
			return &entities.User{
				AccountStatus:     account.StatusActive,
				Password:          hashed,
				PasswordExpiresAt: time.Now().Add(-1 * time.Minute),
			}, nil
		},
	}
	svc := NewUserService(repo, &fakeMailService{})

	err := svc.Login(context.Background(), &dtos.UserLoginDto{
		Email:    "user@example.com",
		Password: "StrongPass123!",
	})

	require.ErrorIs(t, err, ErrExpiredPassword)
}

func TestUserServiceLoginInvalidPassword(t *testing.T) {
	hashed, _ := auth.HashPassword("StrongPass123!")
	repo := &fakeUserRepo{
		findUserByEmailFn: func(ctx context.Context, email string) (*entities.User, error) {
			return &entities.User{
				AccountStatus:     account.StatusActive,
				Password:          hashed,
				PasswordExpiresAt: time.Now().Add(1 * time.Hour),
			}, nil
		},
	}
	svc := NewUserService(repo, &fakeMailService{})

	err := svc.Login(context.Background(), &dtos.UserLoginDto{
		Email:    "user@example.com",
		Password: "WrongPass123!",
	})

	require.ErrorIs(t, err, ErrBadCredentials)
}

func TestUserServiceLoginSuccess(t *testing.T) {
	hashed, _ := auth.HashPassword("StrongPass123!")
	userID := primitive.NewObjectID()
	repo := &fakeUserRepo{
		findUserByEmailFn: func(ctx context.Context, email string) (*entities.User, error) {
			return &entities.User{
				ID:                userID,
				AccountStatus:     account.StatusActive,
				Password:          hashed,
				PasswordExpiresAt: time.Now().Add(1 * time.Hour),
				Email:             "user@example.com",
			}, nil
		},
	}
	mail := &fakeMailService{}
	svc := NewUserService(repo, mail)

	start := time.Now()
	err := svc.Login(context.Background(), &dtos.UserLoginDto{
		Email:    "user@example.com",
		Password: "StrongPass123!",
	})

	require.NoError(t, err)
	require.True(t, repo.setLoginCalled, "expected SetLoginOtp to be called")
	require.True(t, mail.loginOtpCalled, "expected login OTP email to be sent")
	require.Len(t, mail.loginOtpCode, 6, "expected 6 digit OTP")
	for i := range len(mail.loginOtpCode) {
		require.True(t, mail.loginOtpCode[i] >= '0' && mail.loginOtpCode[i] <= '9', "expected numeric OTP")
	}
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(repo.setLoginHash), []byte(mail.loginOtpCode)))
	require.True(t, repo.setLoginExpiry.After(start.Add(4*time.Minute)) &&
		repo.setLoginExpiry.Before(start.Add(6*time.Minute)), "expected expiry around 5 minutes from now")
}

func TestUserServiceVerifyLoginOtpExpired(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	userID := primitive.NewObjectID()
	repo := &fakeUserRepo{
		findUserByEmailFn: func(ctx context.Context, email string) (*entities.User, error) {
			return &entities.User{
				ID:            userID,
				AccountStatus: account.StatusActive,
				OTPCode: entities.OTPCode{
					Hash:   string(hash),
					Expiry: time.Now().Add(-1 * time.Minute),
				},
			}, nil
		},
	}
	svc := NewUserService(repo, &fakeMailService{})

	_, err := svc.VerifyLoginOtp(context.Background(), &dtos.VerifyLoginOtpDto{
		Email: "user@example.com",
		Code:  "123456",
	})

	require.ErrorIs(t, err, ErrOtpExpired)
	require.True(t, repo.clearLoginCalled)
	require.Equal(t, userID, repo.clearLoginUserID, "expected ClearLoginOtp to be called on expiry")
}

func TestUserServiceVerifyLoginOtpInvalid(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	repo := &fakeUserRepo{
		findUserByEmailFn: func(ctx context.Context, email string) (*entities.User, error) {
			return &entities.User{
				AccountStatus: account.StatusActive,
				OTPCode: entities.OTPCode{
					Hash:   string(hash),
					Expiry: time.Now().Add(5 * time.Minute),
				},
			}, nil
		},
	}
	svc := NewUserService(repo, &fakeMailService{})

	_, err := svc.VerifyLoginOtp(context.Background(), &dtos.VerifyLoginOtpDto{
		Email: "user@example.com",
		Code:  "000000",
	})

	require.ErrorIs(t, err, ErrOtpInvalid)
	require.False(t, repo.clearLoginCalled, "expected ClearLoginOtp to not be called for invalid code")
}

func TestUserServiceVerifyLoginOtpSuccess(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	userID := primitive.NewObjectID()
	repo := &fakeUserRepo{
		findUserByEmailFn: func(ctx context.Context, email string) (*entities.User, error) {
			return &entities.User{
				ID:            userID,
				AccountStatus: account.StatusActive,
				OTPCode: entities.OTPCode{
					Hash:   string(hash),
					Expiry: time.Now().Add(5 * time.Minute),
				},
			}, nil
		},
	}
	svc := NewUserService(repo, &fakeMailService{})

	user, err := svc.VerifyLoginOtp(context.Background(), &dtos.VerifyLoginOtpDto{
		Email: "user@example.com",
		Code:  "123456",
	})

	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, userID, user.ID)
	require.True(t, repo.clearLoginCalled)
	require.Equal(t, userID, repo.clearLoginUserID, "expected ClearLoginOtp to be called on success")
}

func TestUserServiceResendLoginOtpNotFound(t *testing.T) {
	repo := &fakeUserRepo{
		findUserByEmailFn: func(ctx context.Context, email string) (*entities.User, error) {
			return nil, errors.New("not found")
		},
	}
	svc := NewUserService(repo, &fakeMailService{})

	err := svc.ResendLoginOtp(context.Background(), "user@example.com")

	require.ErrorIs(t, err, ErrUserNotFound)
}

func TestUserServiceResendLoginOtpInactive(t *testing.T) {
	repo := &fakeUserRepo{
		findUserByEmailFn: func(ctx context.Context, email string) (*entities.User, error) {
			return &entities.User{AccountStatus: account.StatusInactive}, nil
		},
	}
	svc := NewUserService(repo, &fakeMailService{})

	err := svc.ResendLoginOtp(context.Background(), "user@example.com")

	require.ErrorIs(t, err, ErrUserInactive)
}

func TestUserServiceResendLoginOtpSuccess(t *testing.T) {
	userID := primitive.NewObjectID()
	repo := &fakeUserRepo{
		findUserByEmailFn: func(ctx context.Context, email string) (*entities.User, error) {
			return &entities.User{
				ID:                userID,
				AccountStatus:     account.StatusActive,
				PasswordExpiresAt: time.Now().Add(1 * time.Hour),
				Email:             "user@example.com",
			}, nil
		},
	}
	mail := &fakeMailService{}
	svc := NewUserService(repo, mail)

	err := svc.ResendLoginOtp(context.Background(), "user@example.com")

	require.NoError(t, err)
	require.True(t, repo.setLoginCalled, "expected SetLoginOtp to be called")
	require.True(t, mail.loginOtpCalled, "expected SendLoginOtp to be called")
}

func TestUserServiceCreateNewToken(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	keyFile, err := os.CreateTemp(t.TempDir(), "jwt-private-*.pem")
	require.NoError(t, err)
	defer keyFile.Close()

	pemBlock := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}
	require.NoError(t, pem.Encode(keyFile, pemBlock))
	require.NoError(t, keyFile.Sync())

	t.Setenv("JWT_PRIVATE_KEY_PATH", keyFile.Name())

	svc := NewUserService(&fakeUserRepo{}, &fakeMailService{})
	fixed := time.Date(2025, time.January, 2, 15, 4, 5, 0, time.UTC)
	svc.c = clock.NewFixedClock(fixed)
	user := &entities.User{
		ID:            primitive.NewObjectID(),
		FirstName:     "Jane",
		LastName:      "Doe",
		Username:      "jdoe",
		Role:          account.RoleMember,
		AccountStatus: account.StatusActive,
	}

	tokenString, err := svc.CreateNewToken(context.Background(), user)
	require.NoError(t, err)
	require.NotEmpty(t, tokenString)

	parsed, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, func(token *jwt.Token) (any, error) {
		return &key.PublicKey, nil
	}, jwt.WithTimeFunc(func() time.Time { return fixed }))
	require.NoError(t, err)
	require.True(t, parsed.Valid)
	require.Equal(t, jwt.SigningMethodRS256.Alg(), parsed.Method.Alg())

	claims, ok := parsed.Claims.(jwt.MapClaims)
	require.True(t, ok)

	require.Equal(t, "Jane Doe", claims["name"])
	require.Equal(t, "jdoe", claims["username"])
	require.Equal(t, string(account.RoleMember), claims["role"])
	require.Equal(t, string(account.StatusActive), claims["status"])

	sub, ok := claims["sub"].(string)
	if ok {
		require.Equal(t, user.ID.Hex(), sub)
	} else {
		subObj, ok := claims["sub"].(map[string]any)
		require.True(t, ok)
		require.Equal(t, user.ID.Hex(), subObj["$oid"])
	}

	iatRaw, ok := claims["iat"].(float64)
	require.True(t, ok)
	expRaw, ok := claims["exp"].(float64)
	require.True(t, ok)

	iat := time.Unix(int64(iatRaw), 0).UTC()
	exp := time.Unix(int64(expRaw), 0).UTC()

	require.Equal(t, fixed, iat)
	require.Equal(t, exp, fixed.Add(15*time.Minute))
}

// Change Password Tests
func TestChangePasswordInvalidCurrentPassword(t *testing.T) {
	userID := primitive.NewObjectID()
	hashedPassword, _ := auth.HashPassword("ValidPass123!")
	repo := &fakeUserRepo{
		findUserByIDFn: func(ctx context.Context, id primitive.ObjectID) (*entities.User, error) {
			return &entities.User{
				ID:                  id,
				Email:               "user@example.com",
				Password:            hashedPassword,
				AccountStatus:       account.StatusActive,
				PasswordLastChanged: time.Now().Add(-48 * time.Hour), // Changed more than 24 hours ago
			}, nil
		},
	}
	mail := &fakeMailService{}
	svc := NewUserService(repo, mail)

	ctx := contextWithUserID(context.Background(), userID)
	dto := &dtos.ChangePasswordDto{
		CurrentPassword: "WrongPassword123!",
		NewPassword:     "NewValidPass123!",
	}

	err := svc.ChangePassword(ctx, dto)
	require.ErrorIs(t, err, ErrInvalidCurrentPassword)
}

func TestChangePasswordTooFrequent(t *testing.T) {
	userID := primitive.NewObjectID()
	hashedPassword, _ := auth.HashPassword("ValidPass123!")
	repo := &fakeUserRepo{
		findUserByIDFn: func(ctx context.Context, id primitive.ObjectID) (*entities.User, error) {
			return &entities.User{
				ID:                  id,
				Email:               "user@example.com",
				Password:            hashedPassword,
				AccountStatus:       account.StatusActive,
				PasswordLastChanged: time.Now().Add(-12 * time.Hour), // Changed less than 24 hours ago
			}, nil
		},
	}
	mail := &fakeMailService{}
	svc := NewUserService(repo, mail)

	ctx := contextWithUserID(context.Background(), userID)
	dto := &dtos.ChangePasswordDto{
		CurrentPassword: "ValidPass123!",
		NewPassword:     "NewValidPass123!",
	}

	err := svc.ChangePassword(ctx, dto)
	require.ErrorIs(t, err, ErrTooFrequentPasswordChange)
}

func TestChangePasswordSuccess(t *testing.T) {
	userID := primitive.NewObjectID()
	hashedPassword, _ := auth.HashPassword("ValidPass123!")
	setHashPasswordCalled := false
	var setHashPasswordNewHash string

	repo := &fakeUserRepo{
		findUserByIDFn: func(ctx context.Context, id primitive.ObjectID) (*entities.User, error) {
			return &entities.User{
				ID:                  id,
				Email:               "user@example.com",
				Password:            hashedPassword,
				AccountStatus:       account.StatusActive,
				PasswordLastChanged: time.Now().Add(-48 * time.Hour), // Changed more than 24 hours ago
			}, nil
		},
		setHashPasswordFn: func(ctx context.Context, userId primitive.ObjectID, passwordHash string, newTime, expiresAt time.Time) error {
			setHashPasswordCalled = true
			setHashPasswordNewHash = passwordHash
			return nil
		},
	}
	mail := &fakeMailService{}
	svc := NewUserService(repo, mail)

	ctx := contextWithUserID(context.Background(), userID)
	dto := &dtos.ChangePasswordDto{
		CurrentPassword: "ValidPass123!",
		NewPassword:     "NewValidPass123!",
	}

	err := svc.ChangePassword(ctx, dto)
	require.NoError(t, err)
	require.True(t, setHashPasswordCalled, "expected SetHashPassowrd to be called")
	require.NotEmpty(t, setHashPasswordNewHash, "expected new password hash to be set")
	require.NotEqual(t, hashedPassword, setHashPasswordNewHash, "new password hash should be different from old hash")

	// Verify the new hash is valid
	err = auth.CompareHashAndPassword(setHashPasswordNewHash, "NewValidPass123!")
	require.NoError(t, err, "new password should match the hash")
}
