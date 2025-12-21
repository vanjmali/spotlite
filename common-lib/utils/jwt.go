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
// a parameter,.
type parserFunc[T any] func([]byte) (T, error)

// newKeyLoader function represents a generic closure function which takes an environment variable
// and a parser function as parameters, every type will have it's once, value, err instances which
// will cache the keys and errors,.
func newKeyLoader[T any](envVar string, parser parserFunc[T]) func() (T, error) {
	var (
		once  sync.Once
		value T
		err   error
	)

	// if an error occurs the keys couldn't be fetched and the program will panic,
	return func() (T, error) {
		once.Do(func() {
			value, err = loadKey(envVar, parser)
			if err != nil {
				panic(fmt.Sprintf("CRITICAL: Failed to load key from %s: %v", envVar, err))
			}
		})

		return value, err
	}
}

// loadKey is a local function that does the key fetching and parsing, works for both private and public keys,.
func loadKey[T any](envVar string, parser parserFunc[T]) (T, error) {
	var zero T

	path := os.Getenv(envVar)
	if path == "" {
		return zero, fmt.Errorf("%s is not set", envVar)
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return zero, fmt.Errorf("invalid key path: %w", err)
	}

	data, err := os.ReadFile(filepath.Clean(abs)) // #nosec G304
	if err != nil {
		return zero, fmt.Errorf("failed to read key file: %w", err)
	}

	return parser(data)
}

// GetPublicKey is a wrapper function which does the initial public key fetch or fetch the cached one,.
func GetPublicKey() (*rsa.PublicKey, error) {
	return newKeyLoader("JWT_PUBLIC_KEY_PATH", jwt.ParseRSAPublicKeyFromPEM)()
}

// GetPrivateKey is a wrapper function which does the initial private key fetch or fetch the cached one,.
func GetPrivateKey() (*rsa.PrivateKey, error) {
	return newKeyLoader("JWT_PRIVATE_KEY_PATH", jwt.ParseRSAPrivateKeyFromPEM)()
}
