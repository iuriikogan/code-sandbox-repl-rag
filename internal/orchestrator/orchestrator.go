package orchestrator

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"google.golang.org/genai"
	"githuob.com/iuriikogan/code-sandbox-repl-rag/internal/ai"
	"githuob.com/iuriikogan/code-sandbox-repl-rag/internal/python"
)

// Orchestrator manages the main RAG agent loop.
type Orchestrator struct {
	client *ai.Client
	runner python.Runner
}

// New creates a new Orchestrator.
func New(client *ai.Client, runner python.Runner) *Orchestrator {
	return &Orchestrator{
		client: client,
		runner: runner,
	}
}

// Start begins the orchestration process.
func (o *Orchestrator) Start(ctx context.Context, contextFileName string) error {
	systemInstruction := `You are an elite, cost-optimizing Agentic Router using the Gemini 3.1 family.
There is a massive, UNSTRUCTURED dataset saved at 'context.txt'.

Your goal is to extract a comprehensive summary of all FINANCE related events using a TIERED DISCOVERY approach.

### Tiered Discovery Workflow:
1. **Triage (Python)**: Rapidly scan 'context.txt' using regex or keywords to identify relevant sections.
2. **Sub-Agent Triage (Flash-Lite)**: Use rag.run_sub_agent() to have Gemini 3.1 Flash-Lite quickly evaluate if a text block is worth embedding.
3. **Semantic Search (Flash)**: Use rag.get_embedding() ONLY on the high-value blocks found in Triage.
4. **Deep Analysis**: Perform final filtering and clustering in Python.
5. **Synthesis (Pro)**: Return the distilled manifest for final processing.

### RAG Helper API:
- rag.get_embedding(text: str) -> list[float]: Returns the embedding vector for the given text.
- rag.run_sub_agent(instruction: str, chunk: str) -> str: Dispatches a task to a specialized Gemini 3.1 Flash-Lite sub-agent.
- rag.cosine_similarity(v1: list[float], v2: list[float]) -> float: Calculates similarity between vectors.
- rag.finish(output: str): Returns the final compiled chunks/results and terminates.
- rag.get_context_path() -> str: Returns the path to the 'context.txt' file.
`

	systemInstruction += `
Once the Python tool returns the results, YOU (the Orchestrator) will read them, reason over them, and output the final, polished summary.`

	config := &genai.GenerateContentConfig{
		Temperature: genai.Ptr(float32(0.2)),
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{genai.NewPartFromText(systemInstruction)},
		},
	}

	pythonReplTool := &genai.Tool{
		FunctionDeclarations: []*genai.FunctionDeclaration{{
			Name:        "execute_python_script",
			Description: "Executes Python code in a separate process or secure sandbox environment. Communicate via standard output and formatted JSON as instructed.",
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

	slog.Info("Initializing Orchestrator (Gemini Flash) via Native IPC os/exec (Python)...")
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
			fmt.Println("ORCHESTRATOR MESSAGE:")
			fmt.Println("==================================================")
			fmt.Println(part.Text)
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

			if !strings.HasPrefix(output, "Execution finished without returning a 'done' message.") {
				// The script finished and gave us output. Route to final Synthesis using Gemini Pro.
				return o.doFinalSynthesis(ctx, output)
			}

			// If it failed, let the Orchestrator loop try again by seeing the error
			currentPrompt = []genai.Part{
				*genai.NewPartFromFunctionResponse("execute_python_script", map[string]any{
					"output": output,
				}),
			}
		} else {
            // If no function call, we just break
            break
        }
	}

	return nil
}

func (o *Orchestrator) doFinalSynthesis(ctx context.Context, chunks string) error {
	slog.Info("Sending Python Output to Final Synthesis Model (Gemini Pro)...")

	prompt := fmt.Sprintf(`You are the final synthesis agent. 
You have been provided with highly relevant chunks of a dataset extracted via semantic search.
Your goal is to extract a comprehensive summary of all FINANCE related events.

EXTRACTED CHUNKS:
%s

Read them, reason over them, and output the final, polished summary.`, chunks)

	content := &genai.Content{
		Role:  "user",
		Parts: []*genai.Part{genai.NewPartFromText(prompt)},
	}

	config := &genai.GenerateContentConfig{
		Temperature: genai.Ptr(float32(0.2)),
	}

	resp, err := o.client.GenAIClient.Models.GenerateContent(ctx, ai.FinalSynthesisModelName, []*genai.Content{content}, config)
	if err != nil {
		return fmt.Errorf("final synthesis failed: %w", err)
	}

	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil && len(resp.Candidates[0].Content.Parts) > 0 {
		part := resp.Candidates[0].Content.Parts[0]
		if part.Text != "" {
			fmt.Println("\n==================================================")
			fmt.Println("FINAL ORCHESTRATOR SYNTHESIS:")
			fmt.Println("==================================================")
			fmt.Println(part.Text)
			return nil
		}
	}

	return fmt.Errorf("no response from final synthesis model")
}
