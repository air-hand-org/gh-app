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
Also checks that a Client ID is configured, the private-key path points to a
readable regular file, and the key parses as a supported RSA private key
(PKCS#1 or PKCS#8 PEM, using the same parser as authentication).

App configuration uses `GHAPP_CLIENT_ID` and `GHAPP_PRIVATE_KEY_PATH`, with
`--client-id` and `--private-key-path` taking precedence, just as in normal
execution. Flags can be placed before or after `doctor`.

Each configuration check reports `OK`, `FAIL`, or `SKIP` with a reason. Key
parsing is skipped when the file is unavailable; other checks still run.
Exits with status 0 only when all required checks pass, or nonzero otherwise.
Client ID values, private-key paths and contents, JWTs, and installation tokens
are not printed. Checks are offline: no owner, authentication, or network
access is required, and credentials and repository state are not modified.
A successful local check does not verify installation permissions or that the
key belongs to the configured App.
