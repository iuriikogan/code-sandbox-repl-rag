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
        HandleEmbed(ctx context.Context, chunk string) []float32
        HandleBatchEmbed(ctx context.Context, chunks []string) [][]float32
        HandleBatchCall(ctx context.Context, instruction string, chunks []string) []string
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

        Type        string   `json:"type"`

        Instruction string   `json:"instruction,omitempty"`

        Chunk       string   `json:"chunk,omitempty"`

        Chunks      []string `json:"chunks,omitempty"`

        Output      string   `json:"output,omitempty"`

}

type ipcEmbedResponse struct {
	Vector []float32 `json:"vector"`
}

const pythonHelpers = `
import sys, json, math, os, re
from collections import Counter

def ipc_embed(chunk):
    try:
        print(json.dumps({"type": "embed", "chunk": chunk}))
        sys.stdout.flush()
        line = sys.stdin.readline()
        if not line: return None
        return json.loads(line).get("vector")
    except Exception: return None

def ipc_batch_embed(chunks):
    try:
        print(json.dumps({"type": "batch_embed", "chunks": chunks}))
        sys.stdout.flush()
        line = sys.stdin.readline()
        if not line: return []
        return json.loads(line).get("vectors", [])
    except Exception: return []

def ipc_batch_call(instruction, chunks):
    try:
        print(json.dumps({"type": "batch_call", "instruction": instruction, "chunks": chunks}))
        sys.stdout.flush()
        line = sys.stdin.readline()
        if not line: return []
        return json.loads(line).get("results", [])
    except Exception: return []

def cosine_similarity(v1, v2):
    if not v1 or not v2: return 0.0
    dot_product = sum(a * b for a, b in zip(v1, v2))
    mag_v1 = sum(a * a for a in v1) ** 0.5
    mag_v2 = sum(b * b for b in v2) ** 0.5
    if mag_v1 == 0 or mag_v2 == 0: return 0.0
    return dot_product / (mag_v1 * mag_v2)

def tokenize(text):
    return re.findall(r'\w+', text.lower())

class BM25:
    def __init__(self, corpus):
        self.corpus = corpus
        self.doc_len = [len(doc) for doc in corpus]
        self.avg_doc_len = sum(self.doc_len) / len(corpus) if corpus else 0
        self.df = Counter()
        self.idf = {}
        self.term_freqs = []
        for doc in corpus:
            tf = Counter(doc)
            self.term_freqs.append(tf)
            for term in tf:
                self.df[term] += 1
        for term, freq in self.df.items():
            self.idf[term] = math.log(1 + (len(corpus) - freq + 0.5) / (freq + 0.5))

    def get_scores(self, query, k1=1.5, b=0.75):
        scores = []
        for i in range(len(self.corpus)):
            score = 0.0
            doc_len = self.doc_len[i]
            tf = self.term_freqs[i]
            for term in query:
                if term not in tf: continue
                freq = tf[term]
                numerator = self.idf[term] * freq * (k1 + 1)
                denominator = freq + k1 * (1 - b + b * doc_len / self.avg_doc_len)
                score += numerator / denominator
            scores.append(score)
        return scores

def extract_keywords(text_list, exclude_words):
    words = set(re.findall(r'\b[A-Za-z0-9_-]{4,}\b', " ".join(text_list).lower()))
    exclude = set(w.lower() for w in exclude_words)
    return list(words - exclude)

def get_std_dev_outliers(scores, multiplier=1.8):
    if not scores: return []
    mean = sum(scores) / len(scores)
    variance = sum((x - mean) ** 2 for x in scores) / len(scores)
    std_dev = variance ** 0.5
    if std_dev == 0:
        return [i for i, x in enumerate(scores) if x > mean]
    threshold = mean + multiplier * std_dev
    return [i for i, x in enumerate(scores) if x >= threshold]

def update_vector_rocchio(q_vec, rel_vec, alpha=1.0, beta=0.75):
    if not q_vec: return rel_vec
    if not rel_vec: return q_vec
    return [alpha * q + beta * r for q, r in zip(q_vec, rel_vec)]

def rrf_fusion(bm25_scores, cosine_scores, candidates, k=60):
    bm25_ranks = sorted(range(len(bm25_scores)), key=lambda i: bm25_scores[i], reverse=True)
    cosine_ranks = sorted(range(len(cosine_scores)), key=lambda i: cosine_scores[i], reverse=True)
    bm25_rank_map = {idx: rank for rank, idx in enumerate(bm25_ranks)}
    cosine_rank_map = {idx: rank for rank, idx in enumerate(cosine_ranks)}
    rrf_scores = []
    for i in range(len(candidates)):
        bm25_rank = bm25_rank_map[i]
        cosine_rank = cosine_rank_map[i]
        rrf_score = 1.0 / (k + bm25_rank) + 1.0 / (k + cosine_rank)
        rrf_scores.append(rrf_score)
    return rrf_scores

def get_parent_context(all_chunks, child_indices, window=3):
    parent_chunks = []
    for idx in child_indices:
        start = max(0, idx - window)
        end = min(len(all_chunks), idx + window + 1)
        parent_text = "".join(all_chunks[start:end])
        parent_chunks.append(parent_text)
    return parent_chunks
`

// ExecuteScript runs a Python script natively and handles IPC via stdout/stdin.
func (r *LocalRunner) ExecuteScript(ctx context.Context, code string, contextFileName string, handler IPCHandler) (string, error) {
	tmpFile, err := os.CreateTemp("", "script-*.py")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpFile.Name())

	fullCode := pythonHelpers + "\n" + code
	if _, err := tmpFile.WriteString(fullCode); err != nil {
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

	// Read output line by line natively instead of bufio.Scanner to handle massive dynamic tokens

	reader := bufio.NewReader(stdout)

	var fullOutput string

	var doneOutput string

	var doneReceived bool

	for {

		lineBytes, err := reader.ReadBytes('\n')

		if len(lineBytes) > 0 {

			line := string(lineBytes)

			fullOutput += line

			var msg ipcMessage

			if unmarshalErr := json.Unmarshal(lineBytes, &msg); unmarshalErr == nil {

				switch msg.Type {

				case "embed":

					vector := handler.HandleEmbed(ctx, msg.Chunk)

					resp := ipcEmbedResponse{Vector: vector}

					respBytes, _ := json.Marshal(resp)

					fmt.Fprintf(stdin, "%s\n", respBytes)

				                                                                        case "batch_embed":
				
				                                                                                vectors := handler.HandleBatchEmbed(ctx, msg.Chunks)
				
				                                                                                respBytes, _ := json.Marshal(map[string]any{"vectors": vectors})
				
				                                                                                fmt.Fprintf(stdin, "%s\n", respBytes)
				
				                                                                        case "batch_call":
				
				                                                                                results := handler.HandleBatchCall(ctx, msg.Instruction, msg.Chunks)
				
				                                                                                respBytes, _ := json.Marshal(map[string]any{"results": results})
				
				                                                                                fmt.Fprintf(stdin, "%s\n", respBytes)
				
				                                                                        case "done":
					doneOutput = msg.Output

					doneReceived = true

					break // Need to exit loop if done

				}

			}

		}

		if doneReceived {

			break

		}

		if err != nil {

			break // EOF or other read error

		}

	}

	cmd.Wait()

	if !doneReceived {
		resultStr := "Execution finished without returning a 'done' message.\n"
		if len(errBytes) > 0 {
			errStr := string(errBytes)
			if len(errStr) > 5000 {
				errStr = errStr[:5000] + "\n...[TRUNCATED]"
			}
			resultStr += "Standard Error:\n" + errStr
		}
		if fullOutput != "" {
			if len(fullOutput) > 10000 {
				fullOutput = fullOutput[:10000] + "\n...[TRUNCATED]"
			}
			resultStr += "Standard Output:\n" + fullOutput
		}
		slog.Warn(resultStr)
		return resultStr, nil
	}
	return doneOutput, nil
}
