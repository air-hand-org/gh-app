# gh-app

Run GitHub CLI commands with GitHub App installation authentication. Linux
amd64 and arm64 are supported; Windows is not supported.

## Installation

With mise installed, install gh-app from its GitHub releases:

```sh
mise use -g github:air-hand-org/gh-app@latest
```

Install GitHub CLI (`gh`) separately and make it available on `PATH`. Git is
also required when detecting the owner from the current checkout.

## Basic usage

Configure the GitHub App client ID and the path to its private key:

```sh
export GHAPP_CLIENT_ID='your-app-client-id'
export GHAPP_PRIVATE_KEY_PATH='/path/to/private-key.pem'
gh-app --owner OWNER -- repo view OWNER/REPO
```

The App must be installed on the selected owner with permissions for the
requested operation. Put gh-app flags before `--` and GitHub CLI arguments
after it. `--owner` selects the App installation, not the target repository.
You can also set the owner with `GHAPP_OWNER`.

Inside a checkout with a supported GitHub remote, omit the owner to detect it
using Git:

```sh
gh-app -- pr list
```

Normal execution checks for `gh`, plus `git` when no owner is configured,
before resolving the owner or creating a token. Successful checks are silent.

## Check local prerequisites

```sh
gh-app doctor
```

Checks whether `gh` and `git` resolve on `PATH` and prints their resolved paths.
Both checks are reported, even if a dependency is missing. Exits with status 0
when both are found, or nonzero otherwise. No owner, authentication, or network
access is required.
