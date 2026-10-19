package token

import (
	"context"

	"github.com/air-hand-org/gh-app/internal/token/githubapp"
	"github.com/cockroachdb/errors"
	"github.com/google/go-github/v93/github"
	"github.com/jferrl/go-githubauth"
	"golang.org/x/oauth2"
)

//go:generate moq -out gen_token_generator_moq_test.go . InstallationIDGetter
type InstallationIDGetter interface {
	OrganizationInstallationID(ctx context.Context, owner string) (int64, error)
	UserInstallationID(ctx context.Context, owner string) (int64, error)
}

func inspectInstallationID(ctx context.Context, owner string, installation InstallationIDGetter) (int64, error) {
	instID, orgErr := installation.OrganizationInstallationID(ctx, owner)
	if orgErr == nil {
		return instID, nil
	}

	instID, userErr := installation.UserInstallationID(ctx, owner)
	if userErr == nil {
		return instID, nil
	}

	return -1, errors.Join(
		errors.Wrap(orgErr, "get organization installation"),
		errors.Wrap(userErr, "get user installation"),
	)
}

type installationTokenGenerator struct{}

func (g *installationTokenGenerator) New(ctx context.Context, owner string, clientID string, privateKeyPath string) (string, error) {
	if err := ValidateClientID(clientID); err != nil {
		return "", err
	}
	privateKey, err := ReadPrivateKey(privateKeyPath)
	if err != nil {
		return "", err
	}

	appTokenSource, err := NewApplicationTokenSource(
		clientID,
		privateKey,
	)
	if err != nil {
		return "", err
	}

	client, err := github.NewClient(
		github.WithHTTPClient(
			oauth2.NewClient(ctx, appTokenSource),
		),
	)
	if err != nil {
		return "", err
	}

	installationID, err := inspectInstallationID(ctx, owner, githubapp.NewInstallationIDGetter(client))
	if err != nil {
		return "", err
	}

	token, err := githubauth.NewInstallationTokenSource(
		installationID,
		appTokenSource,
		githubauth.WithContext(ctx),
		githubauth.WithRetryOnThrottle(true),
	).Token()
	if err != nil {
		return "", err
	}

	return token.AccessToken, nil
}

func NewInstallationTokenGenerator() *installationTokenGenerator {
	return &installationTokenGenerator{}
}
