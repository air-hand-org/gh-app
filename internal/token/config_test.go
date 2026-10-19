package token

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadPrivateKeySafeErrors(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sensitive-key-path")
	data, err := ReadPrivateKey(path)
	require.Nil(t, data)
	require.ErrorIs(t, err, ErrKeyNotExist)
	require.EqualError(t, err, "file does not exist")
	require.Nil(t, errors.Unwrap(err))
	require.NotContains(t, fmt.Sprintf("%+v", err), path)
}

func TestCheckConfiguration(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	pkcs8, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)
	ec, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	ec8, err := x509.MarshalPKCS8PrivateKey(ec)
	require.NoError(t, err)
	for _, tc := range []struct {
		name  string
		data  []byte
		valid bool
	}{
		{"invalid", []byte("sensitive-key-contents"), false},
		{"ECDSA", pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: ec8}), false},
		{"public", pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: x509.MarshalPKCS1PublicKey(&key.PublicKey)}), false},
		{"PKCS1", pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}), true},
		{"PKCS8", pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8}), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "sensitive-key-path")
			require.NoError(t, os.WriteFile(path, tc.data, 0o600))
			result := CheckConfiguration("", path)
			require.ErrorIs(t, result.ClientIDErr, ErrNotConfigured)
			require.NoError(t, result.KeyFileErr)
			if tc.valid {
				require.NoError(t, result.KeyErr)
			} else {
				require.ErrorIs(t, result.KeyErr, ErrInvalidKey)
				require.Nil(t, errors.Unwrap(result.KeyErr))
				require.NotContains(t, fmt.Sprintf("%+v", result.KeyErr), path)
				require.NotContains(t, fmt.Sprintf("%+v", result.KeyErr), string(tc.data))
			}
			actual, err := os.ReadFile(path)
			require.NoError(t, err)
			require.Equal(t, tc.data, actual)
			result = CheckConfiguration("sensitive-client-id", path)
			require.NoError(t, result.ClientIDErr)
		})
	}
}

func TestConfigurationUnavailableKey(t *testing.T) {
	for _, kind := range []string{"missing config", "nonexistent", "directory", "unreadable", "inaccessible"} {
		t.Run(kind, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "sensitive-key-path")
			expected := ErrKeyNotExist
			switch kind {
			case "missing config":
				path = ""
				expected = ErrNotConfigured
			case "directory":
				require.NoError(t, os.Mkdir(path, 0o700))
				expected = ErrKeyNotRegular
			case "unreadable":
				require.NoError(t, os.WriteFile(path, []byte("sensitive-key-contents"), 0o000))
				expected = ErrKeyNotReadable
			case "inaccessible":
				require.NoError(t, os.Mkdir(path, 0o000))
				path = filepath.Join(path, "key")
				expected = ErrKeyNotAccessible
			}
			result := CheckConfiguration("client", path)
			require.NoError(t, result.ClientIDErr)
			require.ErrorIs(t, result.KeyFileErr, expected)
			require.Nil(t, errors.Unwrap(result.KeyFileErr))
			require.Nil(t, result.KeyErr, "parsing must be skipped")
			if path != "" {
				require.NotContains(t, fmt.Sprintf("%+v", result.KeyFileErr), path)
			}
		})
	}
}
