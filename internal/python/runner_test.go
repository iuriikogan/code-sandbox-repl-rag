package python

import (
	"context"
	"strings"
	"testing"
)

// MockHandler is a simple mock for the IPCHandler interface.
type MockHandler struct {
	CalledHandleBatchCall  bool
	CalledHandleEmbed      bool
	CalledHandleBatchEmbed bool
}

func (m *MockHandler) HandleBatchCall(ctx context.Context, instruction string, chunks []string) []string {
	m.CalledHandleBatchCall = true
	res := make([]string, len(chunks))
	for i := range chunks {
		res[i] = "mock result"
	}
	return res
}

func (m *MockHandler) HandleEmbed(ctx context.Context, chunk string) []float32 {
	m.CalledHandleEmbed = true
	return []float32{1.0, 2.0}
}

func (m *MockHandler) HandleBatchEmbed(ctx context.Context, chunks []string) [][]float32 {
	m.CalledHandleBatchEmbed = true
	res := make([][]float32, len(chunks))
	for i := range chunks {
		res[i] = []float32{1.0, 2.0}
	}
	return res
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

# Test Batch Call
print(json.dumps({"type": "batch_call", "instruction": "test", "chunks": ["data"]}))
sys.stdout.flush()
response_call = json.loads(sys.stdin.readline())

# Test Embed
print(json.dumps({"type": "embed", "chunk": "data"}))
sys.stdout.flush()
response_embed = json.loads(sys.stdin.readline())

# Test Batch Embed
print(json.dumps({"type": "batch_embed", "chunks": ["data"]}))
sys.stdout.flush()
response_batch_embed = json.loads(sys.stdin.readline())

# Return output
results = response_call.get("results", [])
result_str = "".join(results) if results else ""
vector = response_embed.get("vector", [])
print(json.dumps({"type": "done", "output": result_str + str(vector)}))
sys.stdout.flush()
`

	out, err := runner.ExecuteScript(ctx, script, "dummy.txt", handler)
	if err != nil {
		t.Fatalf("ExecuteScript failed: %v", err)
	}

	if !handler.CalledHandleBatchCall {
		t.Errorf("Expected HandleBatchCall to be invoked")
	}
	if !handler.CalledHandleEmbed {
		t.Errorf("Expected HandleEmbed to be invoked")
	}
	if !handler.CalledHandleBatchEmbed {
		t.Errorf("Expected HandleBatchEmbed to be invoked")
	}

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
