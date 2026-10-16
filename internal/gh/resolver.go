package gh

import (
	"context"
	"net/url"
	"os/exec"
	"slices"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/samber/lo/it"
)

const (
	GitScpPrefix string = "git@github.com:"
	GitURLHost   string = "github.com"
)

func parseOwnerFromGitRemote(remoteURL string) (string, error) {
	if strings.HasPrefix(remoteURL, GitScpPrefix) && strings.Count(remoteURL, "/") == 1 {
		remoteURL := strings.TrimPrefix(remoteURL, GitScpPrefix)
		owner, repo, found := strings.Cut(remoteURL, "/")
		if !found || len(owner) == 0 || len(repo) == 0 {
			return "", errors.New("invalid url")
		}
		return owner, nil
	}

	remote, err := url.Parse(remoteURL)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse remoteURL %s", remoteURL)
	}

	if remote.Hostname() != GitURLHost {
		return "", errors.Newf("remote host doesn't match with %s", GitURLHost)
	}

	remotePath := strings.TrimPrefix(remote.EscapedPath(), "/")
	if strings.Count(remotePath, "/") != 1 {
		return "", errors.New("invalid remote path")
	}
	owner, repo, found := strings.Cut(remotePath, "/")
	if !found || len(owner) == 0 || len(repo) == 0 {
		return "", errors.New("invalid url")
	}
	return owner, nil
}

func pickRemoteNameByGitCommand(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "remote")

	remoteName, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, "failed to run git remote")
	}
	remoteLines := it.Map(strings.Lines(string(remoteName)), func(item string) string {
		return strings.TrimSpace(item)
	})
	remoteNames := slices.Collect(it.Filter(remoteLines, func(item string) bool {
		return len(item) != 0
	}))
	if len(remoteNames) != 1 {
		return "", errors.New("failed to resolve remote name")
	}
	return remoteNames[0], nil
}

func resolveRemoteURLByGitCommand(ctx context.Context, remoteName string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "remote", "get-url", remoteName)

	remoteURL, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, "failed to run git remote")
	}
	return strings.TrimSpace(string(remoteURL)), nil
}

type ownerResolver struct{}

func (r *ownerResolver) Owner(ctx context.Context, explicitOwner string) (string, error) {
	if explicitOwner != "" {
		return explicitOwner, nil
	}
	remote, err := pickRemoteNameByGitCommand(ctx)
	if err != nil {
		return "", errors.Wrap(err, "failed to pick remote name")
	}

	remoteURL, err := resolveRemoteURLByGitCommand(ctx, remote)
	if err != nil {
		return "", errors.Wrap(err, "failed to resolve remote url")
	}

	return parseOwnerFromGitRemote(remoteURL)
}

func NewOwnerResolver() *ownerResolver {
	return &ownerResolver{}
}
