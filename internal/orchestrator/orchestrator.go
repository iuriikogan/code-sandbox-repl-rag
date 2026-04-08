package orchestrator

import (
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/genai"
	"github.com/iuriikogan/code-sandbox-repl-rag/internal/ai"
	"github.com/iuriikogan/code-sandbox-repl-rag/internal/python"
	"github.com/iuriikogan/code-sandbox-repl-rag/internal/ui"
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
func (o *Orchestrator) Start(ctx context.Context, contextFileName string, initialPrompt string) error {

	systemInstruction := `You are an elite, cost-optimizing Agentic Router. 

	There is a massive, UNSTRUCTURED dataset saved at 'context.txt' (approx 50MB / 1.2M+ lines).

	Do NOT ask me for the data. Read it from the file 'context.txt'.

	

	Your goal is to extract a comprehensive summary of all highly relevant events.

	

	Because the dataset is massive, you MUST use a Two-Stage Hybrid Search (Lexical + Semantic) to be efficient:

	1. Write a Python script to read the file 'context.txt'.

	2. STAGE 1 (Lexical Filter): Chunk the data. Extract keywords from your target query and use pure Python string matching (or regex) to aggressively filter the millions of lines down to a maximum of 500 candidate chunks.

	3. STAGE 2 (Semantic Search): Use embeddings ONLY on those few hundred candidate chunks.

	

	4. Use the 'google.cloud.aiplatform' SDK to get embeddings for your candidate chunks and your query.

	   Example (Embeddings):

	   from vertexai.language_models import TextEmbeddingModel

	   model = TextEmbeddingModel.from_pretrained("text-embedding-004")

	   embeddings = model.get_embeddings(["candidate one", "candidate two"])

	

	5. Calculate Cosine Similarity locally in Python between the query vector and each candidate chunk's vector.

	6. Dynamically determine the highly relevant chunks (e.g., top 5 to 10).

	7. Return those compiled high-value chunks by printing them clearly to stdout. Do NOT print the 45MB file. Print ONLY the final synthesized RAG chunks.

	`

	

		systemInstruction += `

	Once the Python tool returns the highly relevant chunks, YOU (the Orchestrator) will read them, reason over them, and output the final, polished summary.`

	config := &genai.GenerateContentConfig{
		Temperature: genai.Ptr(float32(0.2)),
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{genai.NewPartFromText(systemInstruction)},
		},
	}

		pythonReplTool := &genai.Tool{

			FunctionDeclarations: []*genai.FunctionDeclaration{{

				Name:        "run_rag_agent_logic",

				Description: "Executes a specialized RAG script in a secure environment. Pass the full Python logic to handle lexical filtering and semantic retrieval.",

				Parameters: &genai.Schema{

					Type: genai.TypeObject,

					Properties: map[string]*genai.Schema{

						"code": {

							Type:        genai.TypeString,

							Description: "The Python code to execute.",

						},

					},

					Required: []string{"code"},

				},

			}},

		}

		config.Tools = []*genai.Tool{pythonReplTool}

	

		slog.Debug("Creating chat session", "model", ai.OrchestratorModelName)

		chat, err := o.client.GenAIClient.Chats.Create(ctx, ai.OrchestratorModelName, config, nil)

		if err != nil {

			return fmt.Errorf("failed to create chat: %w", err)

		}

	

		slog.Info("Orchestrator initialized", "model", ai.OrchestratorModelName)

		return o.sendPromptAndHandleTools(ctx, chat, contextFileName, initialPrompt)

	}

	

	func (o *Orchestrator) sendPromptAndHandleTools(ctx context.Context, session *genai.Chat, contextFileName, prompt string) error {

		var currentPrompt []genai.Part = []genai.Part{*genai.NewPartFromText(prompt)}

	

		for {

			spinner := ui.NewSpinner("Orchestrator is thinking...")

			spinner.Start()

			resp, err := session.SendMessage(ctx, currentPrompt...)

			spinner.Stop("")

			if err != nil {

				return fmt.Errorf("error sending message: %w", err)

			}

	

			if len(resp.Candidates) == 0 {

				slog.Warn("Received empty candidates from Gemini. Breaking loop.")

				break

			}

	

			cand := resp.Candidates[0]

			if cand.Content == nil || len(cand.Content.Parts) == 0 {

				slog.Warn("Received empty content part", "finishReason", cand.FinishReason)

				break

			}

	

			part := cand.Content.Parts[0]

	

			if part.Text != "" {

				slog.Info("Orchestrator thought", "text", part.Text)

			}

	

			if part.FunctionCall != nil && part.FunctionCall.Name == "run_rag_agent_logic" {

				args := part.FunctionCall.Args

				code, ok := args["code"].(string)

				if !ok {

					return fmt.Errorf("invalid 'code' argument in function call")

				}

	

				slog.Info("Executing RAG script in sandbox...")

				output, err := o.runner.ExecuteScript(ctx, code, contextFileName, o.client)

				if err != nil {

					return fmt.Errorf("failed to execute script: %w", err)

				}

	

							slog.Debug("Sandbox execution finished", "outputLength", len(output))

	

				

	

							if err == nil && output != "" {

	

								slog.Info("Goal achieved in sandbox. Starting final synthesis.")

	

								return o.doFinalSynthesis(ctx, output)

	

							}

	

				

	

							slog.Warn("Sandbox logic failed or returned empty output, feeding back into orchestrator.")

	

							currentPrompt = []genai.Part{

	

								*genai.NewPartFromFunctionResponse("run_rag_agent_logic", map[string]any{

	

									"output": output,

	

								}),

	

							}

	

				

			} else if part.FunctionCall != nil {

				slog.Warn("Model attempted to call unknown tool", "name", part.FunctionCall.Name)

				break

			} else {

				slog.Info("No further tool calls. Orchestrator process complete.")

				break

			}

		}

	

		return nil

	}

	

func (o *Orchestrator) doFinalSynthesis(ctx context.Context, chunks string) error {

	prompt := fmt.Sprintf(`You are the final synthesis agent. 

You have been provided with highly relevant chunks of a dataset extracted via semantic search.

Your goal is to extract a comprehensive summary of the specific scenarios requested.



EXTRACTED CHUNKS:

%s



Read them, reason over them, and output the final, polished summary answering the original scenarios.`, chunks)

	content := &genai.Content{

		Role: "user",

		Parts: []*genai.Part{genai.NewPartFromText(prompt)},
	}

	config := &genai.GenerateContentConfig{

		Temperature: genai.Ptr(float32(0.2)),
	}

	spinner := ui.NewSpinner("Synthesizing final answer with Gemini 2.5 Pro...")

	spinner.Start()

	resp, err := o.client.GenAIClient.Models.GenerateContent(ctx, ai.FinalSynthesisModelName, []*genai.Content{content}, config)

	spinner.Stop("")

	if err != nil {

		return fmt.Errorf("final synthesis failed: %w", err)

	}

	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil && len(resp.Candidates[0].Content.Parts) > 0 {

		part := resp.Candidates[0].Content.Parts[0]

		if part.Text != "" {

			slog.Info("FINAL ORCHESTRATOR SYNTHESIS:\n\n" + part.Text)

			return nil

		}

	}

	return fmt.Errorf("no response from final synthesis model")

}
