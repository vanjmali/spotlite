package auth

import (
	"crypto/rsa"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/golang-jwt/jwt/v5"
)

var (
	prvOnce      sync.Once
	cachedPrvKey *rsa.PrivateKey
	errPrvParse  error
)

func GetPrivateKey() (*rsa.PrivateKey, error) {
	prvOnce.Do(func() {
		privateKeyPath := os.Getenv("JWT_PRIVATE_KEY_PATH")
		if privateKeyPath == "" {
			errPrvParse = fmt.Errorf("private key can't be found")
			return
		}

		absPath, err := filepath.Abs(privateKeyPath)
		if err != nil {
			errPrvParse = fmt.Errorf("invalid private key path: %w", err)
			return
		}
		privateKeyPath = filepath.Clean(absPath)

		keyData, err := os.ReadFile(privateKeyPath)
		if err != nil {
			errPrvParse = fmt.Errorf("failed to read key file: %w", err)
			return
		}

		cachedPrvKey, errPrvParse = jwt.ParseRSAPrivateKeyFromPEM(keyData)
	})

	return cachedPrvKey, errPrvParse
}
