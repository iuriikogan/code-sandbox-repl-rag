package ipc

import (
	"encoding/json"
	"testing"
)

func TestMessageJSON(t *testing.T) {
	msg := Message{
		Type:        "call",
		Instruction: "do something",
		Chunk:       "data chunk",
		Vector:      []float32{1.0, 2.0, 3.0},
	}

	bytes, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	var parsed Message
	if err := json.Unmarshal(bytes, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}

	if parsed.Type != msg.Type {
		t.Errorf("Expected Type %q, got %q", msg.Type, parsed.Type)
	}
	if parsed.Instruction != msg.Instruction {
		t.Errorf("Expected Instruction %q, got %q", msg.Instruction, parsed.Instruction)
	}
	if len(parsed.Vector) != 3 || parsed.Vector[0] != 1.0 {
		t.Errorf("Expected Vector to match, got %v", parsed.Vector)
	}
}
