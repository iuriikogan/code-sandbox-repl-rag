package orchestrator

import (
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/genai"
	"githuob.com/iuriikogan/code-sandbox-repl-rag/internal/ai"
	"githuob.com/iuriikogan/code-sandbox-repl-rag/internal/python"
)

// Orchestrator manages the main RAG agent loop.
type Orchestrator struct {
	client *ai.Client
	runner *python.Runner
}

// New creates a new Orchestrator.
func New(client *ai.Client, runner *python.Runner) *Orchestrator {
	return &Orchestrator{
		client: client,
		runner: runner,
	}
}

// Start begins the orchestration process.
func (o *Orchestrator) Start(ctx context.Context, contextFileName string) error {
	config := &genai.GenerateContentConfig{
		Temperature: genai.Ptr(float32(0.2)),
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{genai.NewPartFromText(`You are an elite, cost-optimizing Agentic Router. 
There is a massive, UNSTRUCTURED dataset saved at the file path provided in the 'CONTEXT_FILE' environment variable.
Do NOT ask me for the data. Read it from the file.

Your goal is to extract a comprehensive summary of all FINANCE related events.

Instead of text-based Map-Reduce, you must use Semantic Search (RAG) to find relevant chunks extremely cheaply:
1. Write a Python script (import sys, json, math) to read the file at 'CONTEXT_FILE'.
2. Chunk the unstructured data into logical pieces (e.g., line-by-line).
3. Embed your target query: Request a vector for a query like "finance revenue cost".
   - Print: {"type": "embed", "chunk": "finance revenue cost metrics"}
   - Flush stdout: sys.stdout.flush()
   - Read stdin: json.loads(sys.stdin.readline())["vector"]
4. Iterate through your chunks and get their vectors using the exact same {"type": "embed"} IPC call.
5. Calculate Cosine Similarity locally in Python between the query vector and each chunk's vector.
6. Select the top 3-5 chunks with the highest similarity scores.
7. Return those compiled high-value chunks back to me (the Orchestrator) by printing: {"type": "done", "output": "<compiled_top_chunks>"}
   - Flush stdout immediately: sys.stdout.flush()

Once the Python tool returns the highly relevant chunks, YOU (the Orchestrator) will read them, reason over them, and output the final, polished summary.`)} ,
		},
	}

	pythonReplTool := &genai.Tool{
		FunctionDeclarations: []*genai.FunctionDeclaration{{
			Name:        "execute_python_script",
			Description: "Executes Python code in a separate process. Communicate via JSON over sys.stdout and sys.stdin as instructed.",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"code": {
						Type:        genai.TypeString,
						Description: "Python code to execute.",
					},
				},
				Required: []string{"code"},
			},
		}},
	}
	config.Tools = []*genai.Tool{pythonReplTool}

	chat, err := o.client.GenAIClient.Chats.Create(ctx, ai.OrchestratorModelName, config, nil)
	if err != nil {
		return fmt.Errorf("failed to create chat: %w", err)
	}

	slog.Info("Initializing Orchestrator (Gemini Pro) via Native IPC os/exec (Python)...")
	return o.sendPromptAndHandleTools(ctx, chat, contextFileName, "Begin your task. Inspect the data, chunk it using Python, and dispatch your sub-agents.")
}

func (o *Orchestrator) sendPromptAndHandleTools(ctx context.Context, session *genai.Chat, contextFileName, prompt string) error {
	var currentPrompt []genai.Part = []genai.Part{*genai.NewPartFromText(prompt)}

	for {
		resp, err := session.SendMessage(ctx, currentPrompt...)
		if err != nil {
			return fmt.Errorf("error sending message: %w", err)
		}

		if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
			break
		}

		part := resp.Candidates[0].Content.Parts[0]

		if part.Text != "" {
			fmt.Println("\n==================================================")
			fmt.Println("FINAL ORCHESTRATOR SYNTHESIS:")
			fmt.Println("==================================================")
			fmt.Println(part.Text)
			break
		}

		if part.FunctionCall != nil && part.FunctionCall.Name == "execute_python_script" {
			args := part.FunctionCall.Args
			
			codeAny, ok := args["code"]
			if !ok {
				return fmt.Errorf("missing 'code' argument in function call")
			}
			code, ok := codeAny.(string)
			if !ok {
				return fmt.Errorf("invalid 'code' argument type in function call")
			}
			
			slog.Info("Orchestrator Executing Python via os/exec...")

			output, err := o.runner.ExecuteScript(ctx, code, contextFileName, o.client)
			if err != nil {
				return fmt.Errorf("failed to execute python script: %w", err)
			}

			currentPrompt = []genai.Part{
				*genai.NewPartFromFunctionResponse("execute_python_script", map[string]any{
					"output": output,
				}),
			}
		} else {
            // Nothing to do
            break
        }
	}

	return nil
}