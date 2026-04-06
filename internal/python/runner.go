package python

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"strings"

	"githuob.com/iuriikogan/code-sandbox-repl-rag/internal/ipc"
)

// IPCHandler defines the interface for handling IPC calls from the Python script.
type IPCHandler interface {
	HandleCall(ctx context.Context, instruction, chunk string) string
	HandleEmbed(ctx context.Context, chunk string) []float32
}

// Runner provides methods to execute Python scripts.
type Runner struct{}

// NewRunner creates a new Python runner.
func NewRunner() *Runner {
	return &Runner{}
}

// getPythonCmd locates the Python executable on the system.
func getPythonCmd() string {
	if _, err := exec.LookPath("python3"); err == nil {
		return "python3"
	}
	return "python"
}

// ExecuteScript runs a Python script, connecting it to the provided handler via JSON over stdout/stdin.
func (r *Runner) ExecuteScript(ctx context.Context, code string, contextFileName string, handler IPCHandler) (string, error) {
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

	stdin, err := cmd.StdinPipe()
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

	var finalOutput string
	var debugLogs strings.Builder

	// Use a large buffer to accommodate potentially massive JSON chunks
	scanner := bufio.NewScanner(stdout)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		var msg ipc.Message

		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			// Not a valid JSON IPC message, treat as debug output
			debugLogs.WriteString(line + "\n")
			continue
		}

		if msg.Type == "call" {
			res := handler.HandleCall(ctx, msg.Instruction, msg.Chunk)
			respBytes, _ := json.Marshal(ipc.Message{Result: res})
			stdin.Write(append(respBytes, '\n'))
		} else if msg.Type == "embed" {
			vec := handler.HandleEmbed(ctx, msg.Chunk)
			respBytes, _ := json.Marshal(ipc.Message{Vector: vec})
			stdin.Write(append(respBytes, '\n'))
		} else if msg.Type == "done" {
			finalOutput = msg.Output
		}
	}

	errBytes, _ := io.ReadAll(stderr)
	cmd.Wait() // wait for process to finish

	if finalOutput == "" {
		resultStr := "Execution finished without returning a 'done' message.\n"
		if debugLogs.Len() > 0 {
			resultStr += "Standard Output (Debug):\n" + debugLogs.String() + "\n"
		}
		if len(errBytes) > 0 {
			resultStr += "Standard Error:\n" + string(errBytes)
		}
		return resultStr, nil
	}

	return finalOutput, nil
}
