package token

import (
	"errors"
	"os"

	"github.com/jferrl/go-githubauth"
	"golang.org/x/oauth2"
)

// Configuration errors deliberately own their messages and never wrap filesystem
// or parser errors, which can contain credential paths or key material.
var (
	ErrNotConfigured    = errors.New("not configured")
	ErrKeyNotExist      = errors.New("file does not exist")
	ErrKeyNotAccessible = errors.New("file is not accessible")
	ErrKeyNotRegular    = errors.New("not a regular file")
	ErrKeyNotReadable   = errors.New("file is not readable")
	ErrInvalidKey       = errors.New("invalid or unsupported private key")
)

// ConfigurationChecks contains independent offline results. KeyErr is nil when
// KeyFileErr prevents parsing; callers must report that check as skipped.
type ConfigurationChecks struct {
	ClientIDErr error
	KeyFileErr  error
	KeyErr      error
}

func ValidateClientID(clientID string) error {
	if clientID == "" {
		return ErrNotConfigured
	}
	return nil
}

// NewApplicationTokenSource parses a supported RSA key without generating a JWT
// or making network requests. Parser errors are intentionally not exposed.
func NewApplicationTokenSource(clientID string, key []byte) (oauth2.TokenSource, error) {
	source, err := githubauth.NewApplicationTokenSource(clientID, key)
	if err != nil {
		return nil, ErrInvalidKey
	}
	return source, nil
}

func CheckConfiguration(clientID, privateKeyPath string) ConfigurationChecks {
	result := ConfigurationChecks{ClientIDErr: ValidateClientID(clientID)}
	key, err := ReadPrivateKey(privateKeyPath)
	result.KeyFileErr = err
	if err == nil {
		// A fixed issuer keeps key validation independent of missing Client ID.
		_, result.KeyErr = NewApplicationTokenSource("doctor", key)
	}
	return result
}

func ReadPrivateKey(path string) ([]byte, error) {
	if path == "" {
		return nil, ErrNotConfigured
	}
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrKeyNotExist
	}
	if err != nil {
		return nil, ErrKeyNotAccessible
	}
	if !info.Mode().IsRegular() {
		return nil, ErrKeyNotRegular
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, ErrKeyNotReadable
	}
	return data, nil
}
