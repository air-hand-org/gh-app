package githubapp

import (
	"context"

	"github.com/google/go-github/v93/github"
)

type InstallationIDGetter struct {
	client *github.Client
}

func (i *InstallationIDGetter) OrganizationInstallationID(ctx context.Context, owner string) (int64, error) {
	inst, _, err := i.client.Apps.GetOrganizationInstallation(ctx, owner)
	if err != nil {
		return -1, err
	}
	return inst.GetID(), nil
}

func (i *InstallationIDGetter) UserInstallationID(ctx context.Context, owner string) (int64, error) {
	inst, _, err := i.client.Apps.GetUserInstallation(ctx, owner)
	if err != nil {
		return -1, err
	}
	return inst.GetID(), nil
}

func NewInstallationIDGetter(client *github.Client) *InstallationIDGetter {
	return &InstallationIDGetter{
		client: client,
	}
}
