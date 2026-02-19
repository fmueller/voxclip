package clipboard

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

var ErrUnavailable = errors.New("no clipboard command available")

type commandSpec struct {
	name      string
	args      []string
	asyncFire bool
}

func CopyText(ctx context.Context, value string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	cmdSpec, err := detectCommand()
	if err != nil {
		return err
	}

	copyCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	if cmdSpec.asyncFire {
		return copyWithDetachedCommand(cmdSpec, value)
	}

	cmd := exec.CommandContext(copyCtx, cmdSpec.name, cmdSpec.args...)
	cmd.Stdin = strings.NewReader(value)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard

	if runErr := cmd.Run(); runErr != nil {
		if errors.Is(copyCtx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("copy to clipboard timed out: %w", copyCtx.Err())
		}
		return fmt.Errorf("copy to clipboard: %w", runErr)
	}

	return nil
}

func detectCommand() (commandSpec, error) {
	if runtime.GOOS == "darwin" {
		if _, err := exec.LookPath("pbcopy"); err == nil {
			return commandSpec{name: "pbcopy"}, nil
		}
		return commandSpec{}, ErrUnavailable
	}

	if _, err := exec.LookPath("wl-copy"); err == nil {
		return commandSpec{name: "wl-copy"}, nil
	}

	if _, err := exec.LookPath("xclip"); err == nil {
		return commandSpec{name: "xclip", args: []string{"-selection", "clipboard", "-in", "-silent"}, asyncFire: true}, nil
	}

	return commandSpec{}, ErrUnavailable
}

func copyWithDetachedCommand(spec commandSpec, value string) error {
	cmd := exec.Command(spec.name, spec.args...)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("open clipboard stdin: %w", err)
	}

	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		return fmt.Errorf("start clipboard command: %w", err)
	}

	if _, err := io.WriteString(stdin, value); err != nil {
		_ = stdin.Close()
		_ = cmd.Process.Kill()
		return fmt.Errorf("write clipboard data: %w", err)
	}

	if err := stdin.Close(); err != nil {
		_ = cmd.Process.Kill()
		return fmt.Errorf("close clipboard stdin: %w", err)
	}

	_ = cmd.Process.Release()
	return nil
}
