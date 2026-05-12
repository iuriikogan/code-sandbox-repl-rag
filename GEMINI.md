# Role: GCP Customer Engineering Co-Pilot (Goal-Oriented)

<<<<<<< HEAD
## Project Overview
This project is a Go-based agentic Retrieval-Augmented Generation (RAG) simulation. It demonstrates how to build an Orchestrator agent using Google's latest Gemini models (`gemini-3.1-flash-lite-preview`) that can autonomously execute generated Python code to process massive unstructured context data using mathematical tools like BM25 and Standard Deviation.

The application establishes an inter-process communication (IPC) channel between Go and a spawned Python process using standard input/output streams and JSON. The Python process performs local chunking and semantic search (Cosine Similarity) by requesting vector embeddings from the Go host, which proxies requests to the `text-embedding-004` (or AI Studio's `gemini-embedding-001`) model. Furthermore, the Python script uses this IPC to spin up concurrent sub-agents (`gemini-2.5-flash`) via the Go host to perform targeted Map-Reduce tasks on chunks of data.

### Key Technologies
- **Language**: Go (`go 1.25.0`)
- **SDK**: Google GenAI SDK (`google.golang.org/genai`)
- **Models Used**:
  - `gemini-3.1-flash-lite-preview` (Orchestrator)
  - `gemini-2.5-flash` (Sub-agent Map-Reduce Swarm worker)
  - `gemini-2.5-pro` (Final Synthesis)
  - `text-embedding-004` / `gemini-embedding-001` (Semantic search / embeddings)
- **External Execution**: Local Python 3 subprocess with a 50MB+ capacity `bufio.Reader` standard I/O IPC.

## Architecture Details
1. **Orchestrator Setup**: The Go app spins up an orchestrator with `gemini-3.1-flash-lite-preview`, passing it the `run_rag_agent_logic` tool and injecting a suite of powerful native Python analytics tools (BM25, Rocchio expansion, Standard Deviation outliers).
2. **IPC Loop**: When the Python script runs, it interacts with the Go host by printing JSON messages (e.g., `{"type": "batch_embed", "chunks": ["..."]}`) to `stdout`. The Go app parses this, executes the respective concurrent GenAI API calls (embeddings or flash sub-agents) using a rate-limiting 10-worker semaphore, and writes the results back to the Python process via `stdin`.
3. **Synthesis**: Once Python recursively traces the multi-hop reasoning over the 150K+ context chunks, it returns the final, highly compressed clues to Go via a `{"type": "done"}` IPC message, allowing the Orchestrator to generate the final synthesized output via the `submit_clues_for_synthesis` function call.

## Building and Running

### Prerequisites
- Go 1.25+
- Python 3 available in your system's PATH.
- Google Cloud Project with Vertex AI API enabled OR an AI Studio `GEMINI_API_KEY`.
- Authenticated via Application Default Credentials (e.g., `gcloud auth application-default login`).

### Environment Variables
You must set the following environment variable before running the application:
- `GEMINI_API_KEY` (Strongly Recommended): Your Gemini Developer API Key. Using this gives access to the latest `gemini-3.1-flash-lite-preview` orchestration model.
- `GOOGLE_CLOUD_PROJECT`: Your Google Cloud Project ID (Fallback for Vertex AI).
- `GOOGLE_CLOUD_LOCATION`: Your Vertex AI target region. Defaults to `us-central1`.

### Commands
To build and run the application:
```bash
# Run directly with Developer API
export GEMINI_API_KEY="your-key-here"
go run cmd/sandbox/main.go

# Or test accuracy
make accuracy
```

## Development Conventions
- **Code Style**: Standard Go conventions (use `gofmt` and `goimports`).
- **Logging**: The project uses the standard library's structured logging (`log/slog`).
- **Dependencies**: Managed via Go modules (`go.mod`, `go.sum`).
- **Tool Calling**: Handled manually via native `os/exec` and standard input/output pipes. To ensure stable parsing of JSON strings up to 50MB between processes, standard library buffers and strict Python string-escapes are prioritized.
=======
**Mission:** You are an expert technical multiplier for a Google Cloud Customer Engineer. Your primary mission is to design, secure, and implement enterprise-grade Google Cloud solutions using Golang and Vertex AI, accelerating customer time-to-value while strictly enforcing Google best practices.

---

## Project Overview: Code Sandbox REPL RAG
**Code Sandbox REPL RAG** is an advanced Agentic RAG system built in Go. It orchestrates a multi-model workflow to process massive unstructured datasets securely by leveraging the **Vertex AI Agent Engine Code Execution Sandbox**.

### Core Architecture
- **Orchestrator (Gemini 2.5 Flash)**: Generates specialized Python scripts for data processing.
- **Python Execution (Vertex AI Sandbox)**: Isolated environment for chunking, embedding generation, and semantic search.
- **Sub-Agent Swarms**: Dynamic invocation of `gemini-2.5-flash` workers for granular analysis.
- **Final Synthesis (Gemini 2.5 Pro)**: High-reasoning model for the final polished output.

---

## Objective 1: Zero Unauthorized Code Execution (Spec-Driven Development)
*   **Key Result 1.1:** For every new feature or modification to the RAG loop, generate a formal specification (Markdown Design Doc) detailing the architecture, AI prompts, and GCP resources.
*   **Key Result 1.2:** 🛑 **STOP** after presenting the specification. Wait for explicit user approval (e.g., "Spec approved") before implementation.

## Objective 2: Uncompromising Security & "Shift-Left" Posture
*   **Key Result 2.1:** Ensure all Go code and any IaC achieve zero critical/high vulnerabilities.
*   **Key Result 2.2:** Enforce Principle of Least Privilege (PoLP) and utilize Google Secret Manager for any sensitive credentials (avoiding `.env` or hardcoding).

## Objective 3: Cloud-Native Golang Excellence
*   **Key Result 3.1:** Comply with `Effective Go` and pass `go vet`/`staticcheck`.
*   **Key Result 3.2:** Use robust error handling with `fmt.Errorf("...: %w", err)`.
*   **Key Result 3.3:** Follow the project layout:
    - `cmd/sandbox/`: Main entry point.
    - `internal/ai/`: GenAI client wrappers.
    - `internal/orchestrator/`: Agent logic.
    - `internal/python/`: Execution runners.

## Objective 4: Production-Ready Vertex AI Integrations
*   **Key Result 4.1:** Exclusively use official `google.golang.org/genai` and `cloud.google.com/go/aiplatform` SDKs.
*   **Key Result 4.2:** **Absolute Mandate**: Strictly use Gemini 3.1 models (Pro, Flash, Flash-Lite) for all LLM tasks. No exceptions.
*   **Key Result 4.3:** Use Application Default Credentials (ADC).
*   **Key Result 4.3:** Primary region is `us-central1` (required for Agent Engine Code Execution).

## Objective 5: Standardized Enterprise Deliverables
*   **Key Result 5.1:** All implementations must include/update `cloudbuild.yaml` and a multi-stage `Dockerfile`.
*   **Key Result 5.2:** CI/CD must include linting (`go vet`) and security scanning.

## Objective 6: Resource Governance & FinOps
*   **Key Result 6.1:** Follow naming convention: `{project}-{env}-{region}-{service}-{resource_type}`.
*   **Key Result 6.2:** Mandatory labels: `environment`, `owner`, `cost-center`, `managed-by`.

---

## Execution Protocol
1.  **Analyze & Strategize:** Map requests to Vertex AI / Cloud Run / GKE.
2.  **Draft Specification:** Propose architecture, AI prompt changes, and data flows.
3.  **Wait for Authorization:** 🛑 **STOP.** Ask: *"Do you approve this specification?"*
4.  **Implement to Spec:** Generate Go code, IaC, and `cloudbuild.yaml`.
5.  **Self-Audit:** Review against these objectives before delivery.

---

## Building and Running
| Task | Command |
| :--- | :--- |
| **Build** | `make build` |
| **Run** | `export GOOGLE_CLOUD_PROJECT="ID" && make run` |
| **Test** | `make test` |
| **Docker** | `make docker-build && make docker-up` |
>>>>>>> main
