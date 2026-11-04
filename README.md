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

## Commit identity

```sh
gh-app identity --json
gh-app git-config           # repository-local (default)
gh-app git-config --local   # explicitly repository-local
gh-app git-config --global  # explicitly opt into global configuration
```

`identity` authenticates as the configured App to fetch its slug, then resolves
its public bot account. It prints `login <email>` by default; `--json` returns
`id`, `login`, `name`, and `email`. `name` is the bot login and the email is
`ID+LOGIN@users.noreply.github.com`. Neither command needs `gh` or an installation
owner; App configuration flags and environment variables work as above. They
require network access; the public bot lookup is unauthenticated and subject to
GitHub's public API rate limit. Credentials and tokens are never printed.

`git-config` requires Git and replaces `user.name` and `user.email` in the chosen
scope. Local scope requires a Git repository; `--local` and `--global` are
mutually exclusive. Existing values in that scope are overwritten; other scopes
are unchanged. Authentication failure writes nothing. The two Git config writes
are sequential, not transactional: a write failure can leave a partial update.
This changes future commit attribution, not push authentication or existing
commits.

Check without changing configuration:

```sh
gh-app git-config check
```

`git-config check` compares effective `git config --get user.name` (normal Git
scope precedence) exactly with the authenticated App bot name. It prints
confirmation and exits 0 on a match; an unset/different name, Git error, or
App authentication error exits nonzero. Mismatches show actual and expected names.
It never writes config and does not check email or author/committer environment
overrides. It needs Git and the same App credentials/network as `identity`, not
an installation owner or `gh`. Future pre-tool or Git hooks can use its exit
status as a guard; no hooks are installed by this command.

## Install the agent skill

Print the embedded skill with `gh-app skills`, or install it into an explicitly
selected parent directory containing skills:

```sh
gh-app skills install --skills-dir "./my skills"
```

This creates `./my skills/gh-app/SKILL.md`, including missing directories, and
prints the installed path. Existing files are not overwritten; pass `--force`
to replace one. `--skills-dir` is required. Both skill commands work offline
without App configuration, authentication, `gh`, or `git`. No agent configuration
or agent-specific directory is selected automatically.

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
