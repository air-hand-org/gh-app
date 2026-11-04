package token

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/google/go-github/v93/github"
	"golang.org/x/oauth2"
)

// Identity is the authenticated App's bot account and Git commit identity.
type Identity struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type identityResolver struct{}

func NewIdentityResolver() *identityResolver { return &identityResolver{} }

func (r *identityResolver) Resolve(ctx context.Context, clientID, privateKeyPath string) (Identity, error) {
	if err := ValidateClientID(clientID); err != nil {
		return Identity{}, err
	}
	key, err := ReadPrivateKey(privateKeyPath)
	if err != nil {
		return Identity{}, err
	}
	source, err := NewApplicationTokenSource(clientID, key)
	if err != nil {
		return Identity{}, err
	}
	appClient, err := github.NewClient(github.WithHTTPClient(oauth2.NewClient(ctx, source)))
	if err != nil {
		return Identity{}, err
	}
	// Public bot metadata can be read without authentication. Keep the App JWT
	// scoped to App API requests rather than sending it to the users API.
	publicClient, err := github.NewClient()
	if err != nil {
		return Identity{}, err
	}
	return resolveIdentity(ctx, appClient, publicClient)
}

func resolveIdentity(ctx context.Context, appClient, publicClient *github.Client) (Identity, error) {
	app, _, err := appClient.Apps.Get(ctx, "")
	if err != nil {
		return Identity{}, errors.Wrap(err, "get authenticated App")
	}
	if app.GetSlug() == "" {
		return Identity{}, errors.New("authenticated App has no slug")
	}
	login := app.GetSlug() + "[bot]"
	user, _, err := publicClient.Users.Get(ctx, login)
	if err != nil {
		return Identity{}, errors.Wrap(err, "get App bot account")
	}
	if user.GetID() <= 0 || user.GetLogin() != login || user.GetType() != "Bot" {
		return Identity{}, errors.New("invalid App bot account")
	}
	return Identity{ID: user.GetID(), Login: login, Name: login, Email: fmt.Sprintf("%d+%s@users.noreply.github.com", user.GetID(), login)}, nil
}
