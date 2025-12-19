package utils

import (
	"crypto/rsa"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/golang-jwt/jwt/v5"
)

var (
	pubOnce      sync.Once
	cachedPubKey *rsa.PublicKey
	errPubParse  error
)

// GetPublicKey function is implemented as a singleton so that the key itself isn't read multiple
// times from the disk.
func GetPublicKey() (*rsa.PublicKey, error) {
	pubOnce.Do(func() {
		publicKeyPath := os.Getenv("JWT_PUBLIC_KEY_PATH")
		if publicKeyPath == "" {
			errPubParse = fmt.Errorf("pub. key can't be found")
			return
		}

		absPath, err := filepath.Abs(publicKeyPath)
		if err != nil {
			errPubParse = fmt.Errorf("invalid pub. key path: %w", err)
			return
		}

		cleanPath := filepath.Clean(absPath)
		if strings.Contains(cleanPath, "..") {
			errPubParse = fmt.Errorf("unsafe pub. key relative path detected")
			return
		}

		// #nosec G304
		keyData, err := os.ReadFile(cleanPath)
		if err != nil {
			errPubParse = fmt.Errorf("failed to read pub. key file: %w", err)
			return
		}

		cachedPubKey, errPubParse = jwt.ParseRSAPublicKeyFromPEM(keyData)
	})

	return cachedPubKey, errPubParse
}
