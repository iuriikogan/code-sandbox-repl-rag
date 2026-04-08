package python

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/exec"
)

// IPCHandler defines the interface for handling IPC calls from the Python script.
type IPCHandler interface {
	HandleCall(ctx context.Context, instruction, chunk string) string
	HandleEmbed(ctx context.Context, chunk string) []float32
}

// Runner defines the interface for executing Python scripts.
type Runner interface {
	ExecuteScript(ctx context.Context, code string, contextFileName string, handler IPCHandler) (string, error)
}

// LocalRunner provides methods to execute Python scripts locally without IPC loops.
type LocalRunner struct{}

// NewRunner creates a new local Python runner.
func NewRunner() Runner {
	return &LocalRunner{}
}

func getPythonCmd() string {
	if _, err := exec.LookPath("python3"); err == nil {
		return "python3"
	}
	return "python"
}

// ExecuteScript runs a Python script natively and reads stdout.
func (r *LocalRunner) ExecuteScript(ctx context.Context, code string, contextFileName string, handler IPCHandler) (string, error) {
	tmpFile, err := os.CreateTemp("", "script-*.py")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(code); err != nil {
		return "", err
	}
	tmpFile.Close()

	cmd := exec.CommandContext(ctx, getPythonCmd(), tmpFile.Name())
	cmd.Env = append(os.Environ(), "CONTEXT_FILE="+contextFileName)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}

	if err := cmd.Start(); err != nil {
		return "", err
	}

	outBytes, _ := io.ReadAll(stdout)
	errBytes, _ := io.ReadAll(stderr)
	cmd.Wait()

	outStr := string(outBytes)
	if outStr == "" {
		resultStr := "Execution finished without returning a 'done' message.\n"
		if len(errBytes) > 0 {
			resultStr += "Standard Error:\n" + string(errBytes)
		}
		slog.Warn(resultStr)
		return resultStr, nil
	}

	return outStr, nil
}
