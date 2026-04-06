package ipc

// Message defines the JSON protocol structure over Standard I/O.
type Message struct {
	Type        string    `json:"type"`
	Instruction string    `json:"instruction,omitempty"`
	Chunk       string    `json:"chunk,omitempty"`
	Output      string    `json:"output,omitempty"`
	Result      string    `json:"result,omitempty"`
	Vector      []float32 `json:"vector,omitempty"`
}
