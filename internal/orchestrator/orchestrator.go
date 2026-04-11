package orchestrator

import (
        "context"
        "encoding/json"
        "fmt"
        "log/slog"
                "strings"
        
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
func (o *Orchestrator) Start(ctx context.Context, contextFileName string, initialPrompt string) (string, error) {

        systemInstruction := "You are a Python scripting agent mapping data for an external Swarm.\n" +
                "\n" +
                "There is an ultra-massive UNSTRUCTURED dataset at 'CONTEXT_FILE' (env var).\n" +
                "Your goal is to extract the most relevant chunks that might answer the user's prompt.\n" +
                "Do NOT write step-by-step reasoning or chat. ONLY use the provided tools.\n" +
                "\n" +
                "**INJECTED HELPER FUNCTIONS:**\n" +
                "These are ALREADY in your environment:\n" +
                "- `ipc_embed(chunk: str) -> list[float]`\n" +
                "- `ipc_batch_embed(chunks: list[str]) -> list[list[float]]`\n" +
                "- `cosine_similarity(v1, v2) -> float`\n" +
                "- `tokenize(text: str) -> list[str]`\n" +
                "- `BM25(corpus)` (Class with `get_scores(query) -> list[float]`)\n" +
                "\n" +
                "**YOUR EXECUTION STRATEGY:**\n" +

                        "1. **Sample:** Read the first 100 lines. Notice that lines like 'Rule 44A' and 'FF_NEW_UI' are repetitive NOISE. Signal lines (like Rule 44B or actual error crashes) appear ONLY ONCE in the entire file.\n" +

                        "2. **Lexical Filter (BM25):** The dataset is full of semantically similar noise. Use the `BM25` class or `tf_idf_keyword_filter` to aggressively score and prune the millions of lines down to the top ~100 unique candidates based on highly specific query words (e.g. 'OOM-kill', 'Envoy', 'Rule 44B', 'Alpha').\n" +

                        "3. **Swarm Analysis:** DO NOT just print lexical filter results and stop. You MUST send the top 100 unique outliers to `ipc_batch_call` in the SAME SCRIPT. Instruction: 'Extract specific causal evidence, feature flags, user names, sizes, or error IDs. Ignore background noise status messages.'\n" +

                        "4. **Recursive Trace:** Use clues from the Swarm results (e.g. Rule names, user Alice, feature flag FF_ARCHIVE_SYNC) to run ANOTHER BM25 query and Swarm analysis with the new keywords. You MUST trace at least 2 hops to find the root cause.\n" +

                        "5. **Synthesize:** Compile strictly the extracted unique facts into a concise list.\n"

        config := &genai.GenerateContentConfig{
                Temperature: genai.Ptr(float32(0.0)),
                SystemInstruction: &genai.Content{
                        Parts: []*genai.Part{genai.NewPartFromText(systemInstruction)},
                },
        }

        pythonReplTool := &genai.Tool{
                FunctionDeclarations: []*genai.FunctionDeclaration{{
                        Name:        "run_rag_agent_logic",
                        Description: "Executes a Python script in a secure sandbox. The script MUST return a JSON string array of relevant text chunks via the 'done' IPC message.",
                        Parameters: &genai.Schema{
                                Type: genai.TypeObject,
                                Properties: map[string]*genai.Schema{
                                        "code": {Type: genai.TypeString, Description: "The Python code to execute."},
                                },
                                Required: []string{"code"},
                        },
                }},
        }
        
        synthesisTool := &genai.Tool{
                FunctionDeclarations: []*genai.FunctionDeclaration{{
                        Name:        "submit_clues_for_synthesis",
                        Description: "Sends the compiled clues to the final synthesis agent once you have confidently extracted the root cause from the dataset.",
                        Parameters: &genai.Schema{
                                Type: genai.TypeObject,
                                Properties: map[string]*genai.Schema{
                                        "clues": {Type: genai.TypeString, Description: "The compiled facts and clues to synthesize."},
                                },
                                Required: []string{"clues"},
                        },
                }},
        }

        config.Tools = []*genai.Tool{pythonReplTool, synthesisTool}

        slog.Debug("Creating chat session", "model", ai.OrchestratorModelName)
        chat, err := o.client.GenAIClient.Chats.Create(ctx, ai.OrchestratorModelName, config, nil)
        if err != nil {
                return "", fmt.Errorf("failed to create chat: %w", err)
        }

        slog.Info("Orchestrator initialized", "model", ai.OrchestratorModelName)
        return o.sendPromptAndHandleTools(ctx, chat, contextFileName, initialPrompt)
}

func (o *Orchestrator) sendPromptAndHandleTools(ctx context.Context, session *genai.Chat, contextFileName, prompt string) (string, error) {
        var currentPrompt []genai.Part = []genai.Part{*genai.NewPartFromText(prompt)}

        for {
                spinner := ui.NewSpinner("Orchestrator is thinking...")
                spinner.Start()
                resp, err := session.SendMessage(ctx, currentPrompt...)
                spinner.Stop("")
                if err != nil {
                        return "", fmt.Errorf("error sending message: %w", err)
                }

                if len(resp.Candidates) == 0 {
                        slog.Warn("Received empty candidates from Gemini. Breaking loop.")
                        break
                }

                cand := resp.Candidates[0]
                if cand.Content == nil || len(cand.Content.Parts) == 0 {
                        slog.Warn("Received empty content part", "finishReason", cand.FinishReason)
                        if cand.FinishReason == genai.FinishReasonMalformedFunctionCall {
                                slog.Info("Retrying due to malformed function call...")
                                currentPrompt = []genai.Part{*genai.NewPartFromText("Please try again and generate the Python code using the run_rag_agent_logic tool.")}
                                continue
                        }
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
                                return "", fmt.Errorf("invalid 'code' argument in function call")
                        }

                        slog.Info("Executing RAG script in sandbox...")
                        output, err := o.runner.ExecuteScript(ctx, code, contextFileName, o.client)
                        if err != nil {
                                return "", fmt.Errorf("failed to execute script: %w", err)
                        }

                        if strings.Contains(output, "Execution finished without returning a 'done' message.") {
                                slog.Warn("Sandbox logic failed or returned empty output, feeding back into orchestrator.")
                                currentPrompt = []genai.Part{*genai.NewPartFromFunctionResponse("run_rag_agent_logic", map[string]any{"output": output})}
                                continue
                        }

                        // The output should be a JSON array string. Parse it to get the chunks.
                        importJson := "encoding/json"
                        _ = importJson
                        
                        var chunks []string
                        err = json.Unmarshal([]byte(output), &chunks)
                        if err != nil {
                                slog.Warn("Script did not return a valid JSON string array.", "output", output)
                                currentPrompt = []genai.Part{*genai.NewPartFromFunctionResponse("run_rag_agent_logic", map[string]any{"error": "Output was not a JSON array of strings: " + output})}
                                continue
                        }

                        if len(chunks) == 0 {
                                currentPrompt = []genai.Part{*genai.NewPartFromFunctionResponse("run_rag_agent_logic", map[string]any{"error": "No chunks returned."})}
                                continue
                        }

                        slog.Info(fmt.Sprintf("Sending %d chunks to individual Swarm environments...", len(chunks)))
                        clues := o.client.HandleBatchCall(ctx, "Extract causal entities, feature flags, compliance rules, and root cause evidence from this text. Ignore normal status logs.", chunks)
                        
                        cluesStr := strings.Join(clues, "\n")
                        currentPrompt = []genai.Part{*genai.NewPartFromFunctionResponse("run_rag_agent_logic", map[string]any{"clues": cluesStr})}
                        continue

                } else if part.FunctionCall != nil && part.FunctionCall.Name == "submit_clues_for_synthesis" {
                        args := part.FunctionCall.Args
                        clues, ok := args["clues"].(string)
                        if !ok {
                                return "", fmt.Errorf("invalid 'clues' argument")
                        }
                        
                        slog.Info("Goal achieved via Swarm. Starting final synthesis.")
                        return o.doFinalSynthesis(ctx, clues)

                } else if part.FunctionCall != nil {
                        slog.Warn("Model attempted to call unknown tool", "name", part.FunctionCall.Name)
                        break
                } else {
                        slog.Info("No further tool calls. Orchestrator process complete.")
                        return part.Text, nil
                }
        }
        return "", nil
}
	

func (o *Orchestrator) doFinalSynthesis(ctx context.Context, chunks string) (string, error) {

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

                return "", fmt.Errorf("final synthesis failed: %w", err)

        }

        if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil && len(resp.Candidates[0].Content.Parts) > 0 {

                part := resp.Candidates[0].Content.Parts[0]

                if part.Text != "" {

                        slog.Info("FINAL ORCHESTRATOR SYNTHESIS:\n\n" + part.Text)

                        return part.Text, nil

                }

        }

        return "", fmt.Errorf("no response from final synthesis model")

}
