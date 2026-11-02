package token

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-github/v93/github"
	"github.com/stretchr/testify/require"
)

func TestResolveIdentity(t *testing.T) {
	t.Parallel()
	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v3/app":
			_, _ = w.Write([]byte(`{"id":123,"slug":"actual-app","name":"Display name"}`))
		case "/api/v3/users/actual-app[bot]":
			_, _ = w.Write([]byte(`{"id":456,"login":"actual-app[bot]","type":"Bot"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	client, err := github.NewClient(github.WithEnterpriseURLs(server.URL+"/", server.URL+"/"))
	require.NoError(t, err)
	actual, err := resolveIdentity(t.Context(), client, client)
	require.NoError(t, err)
	require.Equal(t, Identity{ID: 456, Login: "actual-app[bot]", Name: "actual-app[bot]", Email: "456+actual-app[bot]@users.noreply.github.com"}, actual)
	require.Equal(t, []string{"/api/v3/app", "/api/v3/users/actual-app[bot]"}, paths)
}

func TestResolveIdentityRejectsInvalidResponses(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, app, bot string
		status         int
	}{
		{"App unauthorized", `{}`, `{}`, http.StatusUnauthorized},
		{"missing slug", `{"id":123}`, `{}`, http.StatusOK},
		{"missing bot id", `{"slug":"app"}`, `{"login":"app[bot]","type":"Bot"}`, http.StatusOK},
		{"wrong bot login", `{"slug":"app"}`, `{"id":456,"login":"other[bot]","type":"Bot"}`, http.StatusOK},
		{"not a bot", `{"slug":"app"}`, `{"id":456,"login":"app[bot]","type":"User"}`, http.StatusOK},
		{"bot lookup fails", `{"slug":"app"}`, `{}`, http.StatusNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if strings.HasSuffix(r.URL.Path, "/app") {
					if tc.name == "App unauthorized" {
						w.WriteHeader(tc.status)
					}
					_, _ = io.WriteString(w, tc.app)
				} else {
					w.WriteHeader(tc.status)
					_, _ = io.WriteString(w, tc.bot)
				}
			}))
			defer server.Close()
			client, err := github.NewClient(github.WithEnterpriseURLs(server.URL+"/", server.URL+"/"))
			require.NoError(t, err)
			identity, err := resolveIdentity(t.Context(), client, client)
			require.Error(t, err)
			require.Empty(t, identity)
		})
	}
}

type identityTransport func(*http.Request) (*http.Response, error)

func (f identityTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestIdentityResolverAppAuthentication(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	path := filepath.Join(t.TempDir(), "key.pem")
	require.NoError(t, os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}), 0o600))
	previous := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previous })
	var paths []string
	http.DefaultTransport = identityTransport(func(r *http.Request) (*http.Response, error) {
		paths = append(paths, r.URL.Path)
		body := `{"id":123,"slug":"app"}`
		if r.URL.Path == "/app" {
			require.True(t, strings.HasPrefix(r.Header.Get("Authorization"), "Bearer "))
		} else {
			require.Equal(t, "/users/app[bot]", r.URL.Path)
			require.Empty(t, r.Header.Get("Authorization"), "JWT must never reach public user lookup")
			body = `{"id":456,"login":"app[bot]","type":"Bot"}`
		}
		return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
	})
	identity, err := NewIdentityResolver().Resolve(context.Background(), "client", path)
	require.NoError(t, err)
	require.Equal(t, "456+app[bot]@users.noreply.github.com", identity.Email)
	require.Equal(t, []string{"/app", "/users/app[bot]"}, paths)
}

func TestIdentityResolverSafeConfigurationErrors(t *testing.T) {
	t.Parallel()
	resolver := NewIdentityResolver()
	path := filepath.Join(t.TempDir(), "sensitive-key-path")
	identity, err := resolver.Resolve(t.Context(), "", path)
	require.ErrorIs(t, err, ErrNotConfigured)
	require.Empty(t, identity)
	identity, err = resolver.Resolve(t.Context(), "client", path)
	require.ErrorIs(t, err, ErrKeyNotExist)
	require.NotContains(t, err.Error(), path)
	require.Empty(t, identity)
	require.NoError(t, os.WriteFile(path, []byte("sensitive-key"), 0o600))
	identity, err = resolver.Resolve(t.Context(), "client", path)
	require.ErrorIs(t, err, ErrInvalidKey)
	require.NotContains(t, err.Error(), "sensitive-key")
	require.Empty(t, identity)
}
