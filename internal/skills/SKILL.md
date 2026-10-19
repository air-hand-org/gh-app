---
name: gh-app
description: Run GitHub CLI commands with GitHub App authentication.
---

# GitHub App authentication

Use this skill when GitHub operations should use the configured
GitHub App rather than a personal GitHub account.

## Prerequisites

- `gh-app` and `gh` are installed.
- `GHAPP_CLIENT_ID` is configured.
- `GHAPP_PRIVATE_KEY_PATH` points to a readable private key file.

## Usage

Place wrapper flags before `--` and GitHub CLI arguments after it.

Specify the installation owner explicitly when operating outside
a checkout or when the target differs from the checkout's owner:

    gh-app --owner OWNER -- repo view OWNER/REPO
    gh-app --owner OWNER -- pr view NUMBER --repo OWNER/REPO
    gh-app --owner OWNER -- api repos/OWNER/REPO --jq '.full_name'

Inside a checkout with one supported GitHub remote, owner detection
can be used:

    gh-app -- pr list

## Troubleshooting

Use `gh-app doctor` during initial setup or when local configuration changes or
fails. It is not required before every command once the setup is working.

    gh-app doctor

The checks run offline: `gh` and `git` on PATH, Client ID presence, a readable
regular private-key file, and a supported RSA private key. Configuration flags
and environment variables use the same precedence as normal execution.
Required check failures produce a nonzero exit status. Configuration diagnostics
do not print the Client ID, private-key path or contents, or generated tokens.
A successful result does not verify authentication, installation access, or
repository permissions; diagnose those separately with an authorized API read.

## Authentication boundaries

- `--owner` selects the installation owner, not the target repository.
  Specify `--repo` or the API path separately when needed.
- Do not inspect or print the private key or generated tokens.
- If App authentication fails, report the failure.
  Do not silently retry with personal credentials.
- Authentication does not authorize additional actions:
  follow the user's requested scope for writes and destructive operations.

## Verification

Verify reads against the expected repository.
After writes, read back the exact target before reporting success.
