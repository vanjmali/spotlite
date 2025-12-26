package utils

import (
	"crypto/rsa"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/golang-jwt/jwt/v5"
)

// parserFunc is a generic type which represents a function which takes an array of bytes as
// a parameter.
type parserFunc[T any] func([]byte) (T, error)

type keyCacheEntry[T any] struct {
	once  sync.Once
	value T
	err   error
}

// keyCache stores all currently loaded keys.
var keyCache sync.Map

// loadKey function represents a generic closure function which takes an environment variable
// and a parser function as parameters, every type will have it's once, value, err instances which
// will cache the keys and errors.
func loadKey[T any](envVar string, parser parserFunc[T]) (T, error) {
	v, _ := keyCache.LoadOrStore(envVar, &keyCacheEntry[T]{})
	e := v.(*keyCacheEntry[T])

	e.once.Do(func() {
		path := os.Getenv(envVar)
		if path == "" {
			e.err = fmt.Errorf("environment variable %s not set", envVar)
			return
		}

		abs, err := filepath.Abs(path)
		if err != nil {
			e.err = fmt.Errorf("invalid key path: %w", err)
			return
		}

		data, err := os.ReadFile(filepath.Clean(abs)) // #nosec G304
		if err != nil {
			e.err = fmt.Errorf("failed to read key file: %w", err)
			return
		}

		key, err := parser(data)
		if err != nil {
			e.err = fmt.Errorf("failed to parse key: %w", err)
			return
		}

		e.value = key
	})

	return e.value, e.err
}

// GetPublicKey is a wrapper function which does the initial public key fetch or fetch the cached one.
func GetPublicKey() (*rsa.PublicKey, error) {
	return loadKey("JWT_PUBLIC_KEY_PATH", jwt.ParseRSAPublicKeyFromPEM)
}

// GetPrivateKey is a wrapper function which does the initial private key fetch or fetch the cached one.
func GetPrivateKey() (*rsa.PrivateKey, error) {
	return loadKey("JWT_PRIVATE_KEY_PATH", jwt.ParseRSAPrivateKeyFromPEM)
}
