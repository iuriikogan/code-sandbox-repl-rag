package python

import (
	"context"
	"strings"
	"testing"
)

// MockHandler is a simple mock for the IPCHandler interface.
type MockHandler struct {
	CalledHandleCall  bool
	CalledHandleEmbed bool
}

func (m *MockHandler) HandleCall(ctx context.Context, instruction, chunk string) string {
	m.CalledHandleCall = true
	return "mock result"
}

func (m *MockHandler) HandleEmbed(ctx context.Context, chunk string) []float32 {
	m.CalledHandleEmbed = true
	return []float32{1.0, 2.0}
}

func TestGetPythonCmd(t *testing.T) {
	cmd := getPythonCmd()
	if cmd != "python3" && cmd != "python" {
		t.Errorf("Unexpected python command: %s", cmd)
	}
}

func TestExecuteScript(t *testing.T) {
	runner := NewRunner()
	handler := &MockHandler{}
	ctx := context.Background()

	// A simple python script that tests our IPC mechanisms
	script := `
import sys
import json

# Test Call
print(json.dumps({"type": "call", "instruction": "test", "chunk": "data"}))
sys.stdout.flush()
response_call = json.loads(sys.stdin.readline())

# Test Embed
print(json.dumps({"type": "embed", "chunk": "data"}))
sys.stdout.flush()
response_embed = json.loads(sys.stdin.readline())

# Return output
print(json.dumps({"type": "done", "output": response_call.get("result", "") + str(response_embed.get("vector", []))}))
sys.stdout.flush()
`

	out, err := runner.ExecuteScript(ctx, script, "dummy.txt", handler)
	if err != nil {
		t.Fatalf("ExecuteScript failed: %v", err)
	}

	if !handler.CalledHandleCall {
		t.Errorf("Expected HandleCall to be invoked")
	}
	if !handler.CalledHandleEmbed {
		t.Errorf("Expected HandleEmbed to be invoked")
	}

	_ = "mock result[1.0, 2.0]"
	if !strings.Contains(out, "mock result") {
		t.Errorf("Expected output to contain 'mock result', got: %q", out)
	}
}

func TestExecuteScript_NoDone(t *testing.T) {
	runner := NewRunner()
	handler := &MockHandler{}
	ctx := context.Background()

	script := `print("just a standard print")`
	out, err := runner.ExecuteScript(ctx, script, "dummy.txt", handler)
	if err != nil {
		t.Fatalf("ExecuteScript failed: %v", err)
	}

	if !strings.Contains(out, "Execution finished without returning a 'done' message.") {
		t.Errorf("Expected failure message, got: %s", out)
	}
	if !strings.Contains(out, "just a standard print") {
		t.Errorf("Expected debug output to contain 'just a standard print', got: %s", out)
	}
}
