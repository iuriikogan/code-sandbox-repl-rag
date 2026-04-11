# Code Sandbox REPL RAG

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

### Run the Simulation
To run the Go orchestrator pipeline:
```bash
# Using AI Studio Developer API:
export GEMINI_API_KEY="your-api-key"
go run cmd/sandbox/main.go

# Using Google Cloud Vertex AI:
export GOOGLE_CLOUD_PROJECT="your-project-id"
export GOOGLE_CLOUD_LOCATION="us-central1"
go run cmd/sandbox/main.go
```

### Testing & Evaluation
The repository includes an advanced testing suite that compares Single-Endpoint accuracy against the Swarm RAG implementation across multiple problem spaces (Medical and Engineering).

```bash
make accuracy
```

## Project Structure
- `cmd/sandbox/main.go`: Main application entry point.
- `internal/ai/`: Wrappers and clients for Google GenAI interactions, including concurrent Sub-Agent swarm handlers.
- `internal/data/`: Massive dataset generator containing hidden signal lines amid millions of noise lines.
- `internal/orchestrator/`: Primary agent orchestration, IPC tool setup, and testing logic.
- `internal/python/`: Local Python subprocess execution logic, handling JSON I/O channels up to 50MB.
- `internal/ui/`: Terminal spinners and visual feedback.