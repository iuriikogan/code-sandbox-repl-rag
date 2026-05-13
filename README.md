# Code Sandbox REPL RAG (Gemini 3.1)

## Project Overview

This project is a Go-based agentic Retrieval-Augmented Generation (RAG) simulation. It demonstrates how to build an Orchestrator agent using Google's latest Gemini models that can autonomously execute generated Python code to process massive unstructured context data efficiently.

Instead of a standard, static RAG pipeline, an **Agentic Router** writes bespoke code to navigate the dataset.

*Dynamic Data Processing (Agentic RAG)*: The orchestrator autonomously handles semantic chunking, BM25/TF-IDF lexical filtering, text embeddings, and cosine similarity searches tailored dynamically to the prompt.

*Hierarchical Agent Swarms*: The executed Python scripts can dynamically spin up smaller sub-agents via inter-process communication (IPC) to perform highly targeted Map-Reduce tasks on specific data chunks in parallel.

*Strategic Model Routing*: The architecture optimizes performance and cost by using faster models for orchestration and sub-agent map-reduction (`gemini-3.1-flash-lite-preview` / `gemini-2.5-flash`), a specialized embedding model for semantic search (`text-embedding-004` / `gemini-embedding-001`), and a heavier reasoning model (`gemini-2.5-pro`) for final synthesis.

### Benefits

*Massive Scalability*: By processing unstructured data natively via Python scripts and mathematical algorithms (BM25, Standard Deviation, Rocchio Vector Expansion) before ever hitting the LLM context window, the application easily filters massive unstructured datasets that would otherwise hit rate limits or context bounds.

*Highly Contextual Retrieval*: Traditional RAG relies on rigid, predefined chunking strategies that often miss context. This agentic approach writes bespoke logic to navigate the data, drastically improving the relevance of the retrieved context.

*Superior Final Output*: By ensuring that the data is meticulously filtered and processed by sub-agents before being handed to a powerful synthesis model, the final generation achieves a higher degree of accuracy and reasoning quality.

### Key Technologies
- **Language**: Go (`go 1.25.0`)
- **SDK**: Google GenAI SDK (`google.golang.org/genai`)
- **Models Used**:
  - `gemini-3.1-flash-lite-preview` (Orchestrator)
  - `gemini-2.5-flash` (Sub-agent Map-Reduce Worker)
  - `gemini-2.5-pro` (Final Synthesis)
  - `text-embedding-004` / `gemini-embedding-001` (Semantic search / embeddings)
- **Execution Environment**: Local Python Subprocess with Concurrent Go IPC Bridge

## Architecture Details
1. **Context Generation**: Go dynamically creates a 45MB+ ultra-massive dataset simulating deep engineering memory leaks and multi-generational medical diagnostic data.
2. **Orchestrator Setup**: The Go app spins up an orchestrator with `gemini-3.1-flash-lite-preview`, passing it the `run_rag_agent_logic` and `submit_clues_for_synthesis` tools.
3. **Execution (Local IPC Runner)**: The generated Python script is run natively by the Go host. The Python script utilizes pre-injected analytical tools to scan the massive dummy file, score chunks with BM25, request batch embeddings via IPC, and calculate standard deviations to find the needle in the haystack.
4. **Swarm Map-Reduce**: The Python script dispatches the top semantic outliers to a swarm of `gemini-2.5-flash` sub-agents via the Go IPC bridge. The Go host limits concurrency to protect API quotas.
5. **Recursive Vector Expansion**: As the swarm uncovers textual clues, the Python script updates its core query vector using Rocchio Expansion to shift its semantic search towards the newly discovered causal links.
6. **Synthesis**: The Orchestrator receives the parsed, high-value compressed clues and feeds them into `gemini-2.5-pro` for a final, polished reasoning output.

## Prerequisites
- Go 1.25+
- Python 3 & pip
- Google Cloud Project with Vertex AI API enabled OR a Gemini Developer API (AI Studio) key.

### Environment Variables
You must set the following environment variables before running the application:

- `GOOGLE_CLOUD_PROJECT`: Your Google Cloud Project ID (required if using Vertex AI).
- `GEMINI_API_KEY`: (Optional) Your Gemini Developer API Key. If set, this overrides the Vertex AI backend and accesses AI Studio models (giving access to latest preview models like `gemini-3.1-flash-lite-preview`).
- `GOOGLE_CLOUD_LOCATION`: (Optional) Target Vertex AI region. Defaults to `us-central1`.

## Setup & Running

### Command-Line Flags
The application supports custom dataset ingestion and query overrides via CLI flags:

| Flag | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `-dataset` | `string` | `""` | Path to external dataset/log file (if empty, generates synthetic 45MB context) |
| `-prompt` | `string` | `""` | Custom query instruction for the Orchestrator (if empty, uses default multi-scenario prompt) |

### Usage Examples

#### 1. Run Default Simulation
Runs the pipeline against the default 45MB synthetic engineering/medical dataset:
```bash
export GEMINI_API_KEY="your-key-here"
go run cmd/sandbox/main.go
```

#### 2. Run Sample Static Cascading Failure Dataset
Immediately analyze our bundled sample cascading failure dataset:
```bash
go run cmd/sandbox/main.go -dataset=testdata/cascading_failure.jsonl -prompt="Trace the root cause of the Redis memory eviction spike and connection pool exhaustion."
```

#### 3. Ingest External Log File
Analyze any arbitrary enterprise log file, document corpus, or JSONL export:
```bash
go run cmd/sandbox/main.go -dataset=/var/log/nginx/access.log -prompt="Identify all SQL injection attempts."
```

#### 4. Generate Custom Cascading Failure JSONL Dataset
Use our standalone CLI generator to output custom-scaled cascading failure logs to disk via Vertex AI:
```bash
export GOOGLE_CLOUD_PROJECT="your-project-id"
go run cmd/generator/main.go -logs=250 -out=my_enterprise_failure.jsonl
```

#### 5. Run Automated Evaluation Harness
Execute the automated Vertex AI synthetic cascading failure test suite:
```bash
export GOOGLE_CLOUD_PROJECT="your-project-id"
go test -v ./internal/orchestrator -run TestSyntheticCascadingFailureRAG
```

