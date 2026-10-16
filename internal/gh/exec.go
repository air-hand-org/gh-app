package gh

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

type ghcliExecutor struct{}

func (e *ghcliExecutor) Exec(ctx context.Context, token string, args []string) error {
	cmd := exec.CommandContext(ctx, "gh", args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("GH_TOKEN=%s", token))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func NewGHCliExecutor() *ghcliExecutor {
	return &ghcliExecutor{}
}
