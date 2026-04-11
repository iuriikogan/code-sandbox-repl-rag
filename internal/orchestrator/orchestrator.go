package orchestrator

import (

	"context"

	"encoding/json"

	"fmt"

	"log/slog"

	"strings"

	"github.com/iuriikogan/code-sandbox-repl-rag/internal/ai"
	"github.com/iuriikogan/code-sandbox-repl-rag/internal/python"
	"github.com/iuriikogan/code-sandbox-repl-rag/internal/ui"
	"google.golang.org/genai"
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
		                "These are ALREADY in your environment. DO NOT IMPORT external packages (no `numpy`, no `rank_bm25`).:\n" +
		                "- `ipc_embed(chunk: str) -> list[float]`\n" +
		                "- `ipc_batch_embed(chunks: list[str]) -> list[list[float]]`\n" +
		                "- `cosine_similarity(v1, v2) -> float`\n" +
		                "- `tokenize(text: str) -> list[str]`\n" +
		                                                "- `BM25(corpus: list[list[str]])` (Class instantiated with tokenized docs, has `get_scores(query: list[str]) -> list[float]`)\n" +
		                                                "- `extract_keywords(text_list: list[str], exclude: list[str]) -> list[str]`\n" +
		                                                "\n" +
		                                                                                                                "**YOUR EXECUTION STRATEGY:**\n" +
		                                                                                                                "Write a Python script that implements a multi-hop search loop directly. You MUST NOT hallucinate or hardcode keywords. You must extract them dynamically.\n" +
		                                                                                                                "\n" +
		                                                                                                                "**Mandatory Python Loop Blueprint:**\n" +
		                                                                                                                "```python\n" +
		                                                                                                                "import os, json, sys\n" +
		                                                                                                                "with open(os.environ['CONTEXT_FILE'], 'r') as f:\n" +
		                                                                                                                "    all_chunks = f.readlines()\n" +
		                                                                                                                "\n" +
		                                                                                                                "query_words = tokenize('YOUR USER PROMPT HERE')\n" +
		                                                                                                                "query_vec = ipc_embed('YOUR USER PROMPT HERE')\n" +
		                                                                                                                "evidence = []\n" +
		                                                                                                                "tokenized_corpus = [tokenize(c) for c in all_chunks]\n" +
		                                                                                                                "bm25 = BM25(tokenized_corpus)\n" +
		                                                                                                                "\n" +
		                                                                                                                "for hop in range(3):\n" +
		                                                                                                                "    # 1. Lexical BM25 Pruning\n" +
		                                                                                                                "    scores = bm25.get_scores(query_words)\n" +
		                                                                                                                "    top_indices = sorted(range(len(scores)), key=lambda i: scores[i], reverse=True)[:100]\n" +
		                                                                                                                "    candidates = [all_chunks[i] for i in top_indices if scores[i] > 0]\n" +
		                                                                                                                "    if not candidates: break\n" +
		                                                                                                                "\n" +
		                                                                                                                "    # 2. Semantic Rank\n" +
		                                                                                                                "    candidate_vecs = ipc_batch_embed(candidates)\n" +
		                                                                                                                "    sim_scores = [cosine_similarity(query_vec, v) for v in candidate_vecs]\n" +
		                                                                                                                "    outlier_indices = get_std_dev_outliers(sim_scores, multiplier=1.8)\n" +
		                                                                                                                "    outliers = [candidates[i] for i in outlier_indices]\n" +
		                                                                                                                "    if not outliers: break\n" +
		                                                                                                                "\n" +
		                                                                                                                "    # 3. Swarm Analysis (Extract Facts & IDs)\n" +
		                                                                                                                "    clues = ipc_batch_call('Extract concrete IDs, flags, rules, and causal facts. Ignore noise.', outliers)\n" +
		                                                                                                                "    evidence.extend([c for c in clues if c.strip()])\n" +
		                                                                                                                "\n" +
		                                                                                                                "    # 4. Recursive Update\n" +
		                                                                                                                "    new_words = extract_keywords(clues, query_words)\n" +
		                                                                                                                "    if not new_words: break\n" +
		                                                                                                                "    query_words.extend(new_words)\n" +
		                                                                                                                "    query_vec = update_vector_rocchio(query_vec, ipc_embed(' '.join(new_words)))\n" +
		                                                                                                                "\n" +
		                                                                                                                "print(json.dumps({\"type\": \"done\", \"output\": '\\n'.join(evidence)}))\n" +
		                                                                                                                "sys.stdout.flush()\n" +
		                                                                                                                "```\n"
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
	
	                
	
	                slog.Debug("Creating chat session", "model", o.client.OrchestratorModelName)
	chat, err := o.client.GenAIClient.Chats.Create(ctx, o.client.OrchestratorModelName, config, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create chat: %w", err)
	}

	slog.Info("Orchestrator initialized", "model", o.client.OrchestratorModelName)
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
			                                                                                                                        
			                                                                                                                        var validClues []string
			                                                                                                                        for _, c := range clues {
			                                                                                                                                if strings.TrimSpace(c) != "" && !strings.Contains(strings.ToLower(c), "no specific") && !strings.Contains(strings.ToLower(c), "ignore") {
			                                                                                                                                        validClues = append(validClues, c)
			                                                                                                                                }
			                                                                                                                        }
			                                                
			                                                                                                                        cluesStr := strings.Join(validClues, "\n")
			                                                                                                                        if cluesStr == "" {
			                                                                                                                                cluesStr = "No relevant clues found in those chunks."
			                                                                                                                        }
			                                                                                                                        
			                                                                                                                        slog.Info("Returning Swarm clues to Orchestrator to plan next hop...")
			                                                                                                                        currentPrompt = []genai.Part{*genai.NewPartFromFunctionResponse("run_rag_agent_logic", map[string]any{"extracted_clues": cluesStr})}
			                                                                                                                        continue
			                                                
			                                                                                                                } else if part.FunctionCall != nil && part.FunctionCall.Name == "submit_clues_for_synthesis" {
			                                                                                                                        args := part.FunctionCall.Args
			                                                                                                                        clues, ok := args["clues"].(string)
			                                                                                                                        if !ok {
			                                                                                                                                return "", fmt.Errorf("invalid 'clues' argument")
			                                                                                                                        }
			                                                                                                                        
			                                                                                                                        slog.Info("Goal achieved via Swarm. Starting final synthesis.")
			                                                                                                                        return o.doFinalSynthesis(ctx, clues)
			                                                
			                                                                                                                } else if part.FunctionCall != nil {			slog.Warn("Model attempted to call unknown tool", "name", part.FunctionCall.Name)
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
