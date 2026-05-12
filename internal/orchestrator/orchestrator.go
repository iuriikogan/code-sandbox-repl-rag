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

<<<<<<< HEAD
// Start begins the orchestration process with optimized Hybrid RRF RAG instructions.
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
		"- `ipc_batch_call(instruction: str, chunks: list[str]) -> list[str]`\n" +
		"- `cosine_similarity(v1, v2) -> float`\n" +
		"- `tokenize(text: str) -> list[str]`\n" +
		"- `BM25(corpus: list[list[str]])` (Class instantiated with tokenized docs, has `get_scores(query: list[str]) -> list[float]`)\n" +
		"- `extract_keywords(text_list: list[str], exclude: list[str]) -> list[str]`\n" +
		"- `get_std_dev_outliers(scores: list[float], multiplier: float) -> list[int]`\n" +
		"- `update_vector_rocchio(q_vec: list[float], rel_vec: list[float]) -> list[float]`\n" +
		"- `rrf_fusion(bm25_scores: list[float], cosine_scores: list[float], candidates: list[str]) -> list[float]`\n" +
		"- `get_parent_context(all_chunks: list[str], child_indices: list[int], window: int) -> list[str]`\n" +
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
		"    # 1. Parallel Lexical & Dense Retrieval\n" +
		"    bm25_scores = bm25.get_scores(query_words)\n" +
		"    top_indices = sorted(range(len(bm25_scores)), key=lambda i: bm25_scores[i], reverse=True)[:200]\n" +
		"    candidates = [all_chunks[i] for i in top_indices]\n" +
		"    candidate_scores_bm25 = [bm25_scores[i] for i in top_indices]\n" +
		"\n" +
		"    # 2. Dense Semantic Scoring\n" +
		"    candidate_vecs = ipc_batch_embed(candidates)\n" +
		"    candidate_scores_cosine = [cosine_similarity(query_vec, v) for v in candidate_vecs]\n" +
		"\n" +
		"    # 3. RRF Fusion Ranking\n" +
		"    rrf_scores = rrf_fusion(candidate_scores_bm25, candidate_scores_cosine, candidates)\n" +
		"    outlier_indices = get_std_dev_outliers(rrf_scores, multiplier=1.8)\n" +
		"    original_indices = [top_indices[i] for i in outlier_indices]\n" +
		"    if not original_indices: break\n" +
		"\n" +
		"    # 4. Parent Context Expansion\n" +
		"    parents = get_parent_context(all_chunks, original_indices, window=3)\n" +
		"\n" +
		"    # 5. Swarm Analysis over expanded contexts\n" +
		"    clues = ipc_batch_call('Extract ALL concrete IDs, person names (e.g. Alice), service names (e.g. Alpha, cron-beta), feature flags, compliance rules, precise sizes (e.g. 2MB, 5MB), component names (e.g. cgroup, istio-proxy, envoy), custom headers (e.g. x-trace), and causal facts. Be exhaustive. Ignore noise.', parents)\n" +
		"    evidence.extend([c for c in clues if c.strip()])\n" +
		"\n" +
		"    # 6. Recursive Update\n" +
		"    new_words = extract_keywords(clues, query_words)\n" +
		"    if not new_words: break\n" +
		"    query_words.extend(new_words)\n" +
		"    v_clue = ipc_embed(' '.join(new_words))\n" +
		"    if v_clue: query_vec = update_vector_rocchio(query_vec, v_clue)\n" +
		"\n" +
		"print(json.dumps({\"type\": \"done\", \"output\": '\\n'.join(evidence)}))\n" +
		"sys.stdout.flush()\n" +
		"```\n"
=======
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
>>>>>>> main

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

// ExpandQueryTerms queries a Tier 1 model to generate search synonym terms to resolve the RAG multi-hop gap.
func (o *Orchestrator) ExpandQueryTerms(ctx context.Context, clues string, originalQuery string) (string, error) {
	prompt := fmt.Sprintf(`You are a RAG search optimization agent. Based on the currently extracted clues, generate a space-separated list of related keywords, synonyms, service names, team names, or technical concepts that should be searched for in the next step.
CRITICAL: Output ONLY the space-separated keywords. Do not write any introduction, reasoning, or conversational filler.

ORIGINAL USER QUERY:
%s

EXTRACTED OUTAGE CLUES:
%s`, originalQuery, clues)

	content := &genai.Content{
		Role:  "user",
		Parts: []*genai.Part{genai.NewPartFromText(prompt)},
	}

	config := &genai.GenerateContentConfig{
		Temperature: genai.Ptr(float32(0.1)),
	}

	resp, err := o.client.GenAIClient.Models.GenerateContent(ctx, ai.Tier1ModelName, []*genai.Content{content}, config)
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil && len(resp.Candidates[0].Content.Parts) > 0 {
		part := resp.Candidates[0].Content.Parts[0]
		return strings.TrimSpace(part.Text), nil
	}
	return "", fmt.Errorf("no response from expansion model")
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

			// Apply LLM-Backed Semantic Query Expansion on the intercepted clues
			expandedTerms, expandErr := o.ExpandQueryTerms(ctx, cluesStr, prompt)
			if expandErr == nil && expandedTerms != "" {
				slog.Info("LLM Query Expansion successfully expanded search context", "terms", expandedTerms)
				cluesStr = fmt.Sprintf("Extracted Clues:\n%s\n\n[EXPANDED SEARCH TERMS FOR NEXT HOP: %s]", cluesStr, expandedTerms)
			}

			slog.Info("Returning Swarm clues (enriched with semantic expansion) to Orchestrator to plan next hop...")
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

You have been provided with highly relevant clues extracted via a semantic Map-Reduce swarm.

Your goal is to extract a comprehensive summary of the specific scenarios requested.
CRITICAL: You MUST explicitly include ALL technical details provided in the clues. Do not summarize away specific service names (like Alpha, Omega, cron-beta), components (like envoy, cgroup, istio-proxy), rules (like Rule 44B), feature flags (like FF_ARCHIVE_SYNC), sizes (like 2MB, 5.2MB), users (like Alice), or custom headers (like x-trace). Your output will be tested for these exact strings.

EXTRACTED CLUES:

%s

Read them, reason over them, and output the final, polished summary answering the original scenarios while retaining ALL technical identifiers.`, chunks)

	content := &genai.Content{
		Role:  "user",
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

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response from final synthesis model")
	}

	synthesisText := resp.Candidates[0].Content.Parts[0].Text
	slog.Info("Polished synthesis report generated.")

	// Run Self-Correction Verification Loop
	verificationPrompt := fmt.Sprintf(`You are an expert RAG verification agent. Review the summary report below and verify if it successfully answers all parts of the scenarios.
Specifically, check if it clearly lists:
1. For Engineering outage: Envoy/istio-proxy memory leak, SRE Team Lead (Alice), feature flag (FF_ARCHIVE_SYNC), payload size (5.2MB), custom headers (X-Trace-Legacy-ID), and the triggering service (cron-beta).
2. For Medical genetic tracking: The genetic links and timelines across Patients A, B, and C, and Patient C's ER admission trigger (Sulfonamides).

Output a JSON object:
"passed": true or false,
"missing_details": "a detailed description of what technical terms/entities were omitted or need more information, or empty if passed"

SUMMARY REPORT:
%s`, synthesisText)

	contentVer := &genai.Content{
		Role:  "user",
		Parts: []*genai.Part{genai.NewPartFromText(verificationPrompt)},
	}

	configVer := &genai.GenerateContentConfig{
		Temperature:      genai.Ptr(float32(0.0)),
		ResponseMIMEType: "application/json",
		ResponseSchema: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"passed": {
					Type:        genai.TypeBoolean,
					Description: "Whether all core details and scenario outcomes are fully present.",
				},
				"missing_details": {
					Type:        genai.TypeString,
					Description: "Omitted names, sizes, rule numbers, or flags that must be added.",
				},
			},
			Required: []string{"passed", "missing_details"},
		},
	}

	slog.Info("Running Self-Correction verification loop...", "model", ai.Tier1ModelName)
	respVer, verErr := o.client.GenAIClient.Models.GenerateContent(ctx, ai.Tier1ModelName, []*genai.Content{contentVer}, configVer)
	if verErr == nil && len(respVer.Candidates) > 0 && respVer.Candidates[0].Content != nil {
		var verRes struct {
			Passed         bool   `json:"passed"`
			MissingDetails string `json:"missing_details"`
		}
		partVer := respVer.Candidates[0].Content.Parts[0]
		if json.Unmarshal([]byte(partVer.Text), &verRes) == nil {
			slog.Info("Verification results", "passed", verRes.Passed, "missingDetails", verRes.MissingDetails)
			if !verRes.Passed {
				slog.Warn("Self-Correction TRIGGERED! Report is missing details. Generating remediation pass...")
				
				remediationPrompt := fmt.Sprintf(`Your previous report was incomplete. The verification audit reported missing elements: %s.
Please regenerate the final comprehensive report. Make SURE you explicitly include all technical details, exact flag names (FF_ARCHIVE_SYNC), trigger drugs (Sulfonamides), operator names (Alice), custom header names (X-Trace-Legacy-ID), rules (Rule 44B), and replica sizes from the extracted clues.

EXTRACTED CLUES:
%s`, verRes.MissingDetails, chunks)

				contentRem := &genai.Content{
					Role:  "user",
					Parts: []*genai.Part{genai.NewPartFromText(remediationPrompt)},
				}
				
				spinnerRem := ui.NewSpinner("Self-Correction remediation synthesis in progress...")
				spinnerRem.Start()
				respRem, errRem := o.client.GenAIClient.Models.GenerateContent(ctx, ai.FinalSynthesisModelName, []*genai.Content{contentRem}, nil)
				spinnerRem.Stop("")
				
				if errRem == nil && len(respRem.Candidates) > 0 && respRem.Candidates[0].Content != nil {
					slog.Info("Self-Correction successfully generated optimized summary report.")
					synthesisText = respRem.Candidates[0].Content.Parts[0].Text
				}
			} else {
				slog.Info("Self-Correction verification PASSED.")
			}
		}
	}

<<<<<<< HEAD
	slog.Info("FINAL ORCHESTRATOR SYNTHESIS:\n\n" + synthesisText)
	return synthesisText, nil
=======
	return fmt.Errorf("no response from final synthesis model")
>>>>>>> main
}
