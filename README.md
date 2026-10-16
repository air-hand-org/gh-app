# gh-app

## Check local prerequisites

```sh
gh-app doctor
```

Checks whether `gh` and `git` resolve on `PATH` and prints their resolved paths.
Both checks are reported, even if a dependency is missing. Exits with status 0
when both are found, or nonzero otherwise. No owner, authentication, or network
access is required.
