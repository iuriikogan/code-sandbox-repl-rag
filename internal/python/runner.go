package python

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
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

// LocalRunner provides methods to execute Python scripts locally with IPC loops.
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

type ipcMessage struct {
	Type        string `json:"type"`
	Instruction string `json:"instruction,omitempty"`
	Chunk       string `json:"chunk,omitempty"`
	Output      string `json:"output,omitempty"`
}

type ipcEmbedResponse struct {
	Vector []float32 `json:"vector"`
}

type ipcCallResponse struct {
	Result string `json:"result"`
}

// ExecuteScript runs a Python script natively and handles IPC via stdout/stdin.
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

	// Capture stderr in a goroutine
	var errBytes []byte
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stderr.Read(buf)
			if n > 0 {
				errBytes = append(errBytes, buf[:n]...)
			}
			if err != nil {
				break
			}
		}
	}()

	scanner := bufio.NewScanner(stdout)
	const maxCapacity = 10 * 1024 * 1024 // 10MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	var fullOutput string
	var doneOutput string
	var doneReceived bool

	for scanner.Scan() {
		line := scanner.Text()
		fullOutput += line + "\n"

		var msg ipcMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			// Not a valid IPC message, ignore or log
			continue
		}

		switch msg.Type {
		case "embed":
			vector := handler.HandleEmbed(ctx, msg.Chunk)
			resp := ipcEmbedResponse{Vector: vector}
			respBytes, _ := json.Marshal(resp)
			fmt.Fprintf(stdin, "%s\n", respBytes)
		case "call":
			result := handler.HandleCall(ctx, msg.Instruction, msg.Chunk)
			resp := ipcCallResponse{Result: result}
			respBytes, _ := json.Marshal(resp)
			fmt.Fprintf(stdin, "%s\n", respBytes)
		case "done":
			doneOutput = msg.Output
			doneReceived = true
			break // exit scanner loop
		}
	}

	if err := scanner.Err(); err != nil {
		slog.Error("Scanner error", "error", err)
	}

	cmd.Wait()

	if !doneReceived {
		resultStr := "Execution finished without returning a 'done' message.\n"
		if len(errBytes) > 0 {
			resultStr += "Standard Error:\n" + string(errBytes)
		}
		if fullOutput != "" {
			resultStr += "Standard Output:\n" + fullOutput
		}
		slog.Warn(resultStr)
		return resultStr, nil
	}

	return doneOutput, nil
}