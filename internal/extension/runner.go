package extension

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/gouzil/yunxiao-cli/internal/terminal"
)

type RealGitRunner struct{}

func (RealGitRunner) Run(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return string(output), nil
}

type RealExecRunner struct {
	GOOS      string
	FindShell func() (string, error)
}

func (r RealExecRunner) Run(ctx context.Context, executable string, args []string, env []string, streams terminal.IOStreams) error {
	cmd := exec.CommandContext(ctx, executable, args...)
	if r.GOOS == "windows" && !strings.HasSuffix(strings.ToLower(executable), ".exe") {
		shell, err := r.shell()
		if err != nil {
			return err
		}
		forwarded := append([]string{"-c", `command "$@"`, "--", executable}, args...)
		cmd = exec.CommandContext(ctx, shell, forwarded...)
	}
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdin = streams.In
	cmd.Stdout = streams.Out
	cmd.Stderr = streams.ErrOut
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return &ExitError{Code: exitErr.ExitCode(), Err: err}
		}
		return err
	}
	return nil
}

func (r RealExecRunner) shell() (string, error) {
	if r.FindShell == nil {
		return defaultFindShell()
	}
	return r.FindShell()
}

func defaultFindShell() (string, error) {
	path, err := exec.LookPath("sh.exe")
	if err != nil {
		return "", fmt.Errorf("the sh.exe interpreter is required. Install Git for Windows and try again")
	}
	return path, nil
}
