package auth

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
	prvOnce      sync.Once
	cachedPrvKey *rsa.PrivateKey
	errPrvParse  error
)

// GetPrivateKey function is implemented as a singleton so that the key itself isn't read multiple
// times from the disk.
func GetPrivateKey() (*rsa.PrivateKey, error) {
	prvOnce.Do(func() {
		privateKeyPath := os.Getenv("JWT_PRIVATE_KEY_PATH")
		if privateKeyPath == "" {
			errPrvParse = fmt.Errorf("priv. key can't be found")
			return
		}

		absPath, err := filepath.Abs(privateKeyPath)
		if err != nil {
			errPrvParse = fmt.Errorf("invalid priv. key path: %w", err)
			return
		}

		cleanPath := filepath.Clean(absPath)
		if strings.Contains(cleanPath, "..") {
			errPrvParse = fmt.Errorf("unsafe priv. key relative path detected")
			return
		}

		// #nosec G304
		keyData, err := os.ReadFile(cleanPath)
		if err != nil {
			errPrvParse = fmt.Errorf("failed to read priv. key file: %w", err)
			return
		}

		cachedPrvKey, errPrvParse = jwt.ParseRSAPrivateKeyFromPEM(keyData)
	})

	return cachedPrvKey, errPrvParse
}
