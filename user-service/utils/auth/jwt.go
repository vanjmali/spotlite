package auth

import (
	"crypto/rsa"
	"fmt"
	"os"
	"sync"

	"github.com/golang-jwt/jwt/v5"
)

var (
	prvOnce      sync.Once
	cachedPrvKey *rsa.PrivateKey
	prvParseErr  error

	pubOnce      sync.Once
	cachedPubKey *rsa.PublicKey
	pubParseErr  error
)

func GetPrivateKey() (*rsa.PrivateKey, error) {
	prvOnce.Do(func() {
		privateKeyPath := os.Getenv("JWT_PRIVATE_KEY_PATH")

		keyData, err := os.ReadFile(privateKeyPath)
		if err != nil {
			prvParseErr = fmt.Errorf("failed to read key file: %w", err)
			return
		}

		cachedPrvKey, prvParseErr = jwt.ParseRSAPrivateKeyFromPEM(keyData)
	})

	return cachedPrvKey, prvParseErr
}
