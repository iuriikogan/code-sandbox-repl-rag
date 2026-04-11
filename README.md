# Code Sandbox REPL RAG

## Project Overview
*Secure, Isolated Execution* By leveraging the Vertex AI Agent Engine Code Execution Sandbox, the system safely runs AI-generated Python code in a secure cloud environment, preventing malicious or flawed code from affecting the host machine.

*Dynamic Data Processing (Agentic RAG)* Instead of relying on a static data pipeline, an orchestrator agent autonomously writes custom code to handle chunking, embedding generation, and cosine similarity searches specifically tailored to the immediate dataset.

*Hierarchical Agent Swarms* The executed Python scripts can dynamically spin up smaller sub-agents within the sandbox to perform highly targeted tasks on specific data chunks in parallel.

*Strategic Model Routing* The architecture optimizes performance and cost by using faster models (Gemini Flash) for orchestration and sub-agent work, a specialized embedding model for semantic search, and a heavier reasoning model (Gemini Pro) for final synthesis.

### Benefits

*Uncompromised Security for Code Generation*: Because LLMs can sometimes hallucinate destructive commands, sandboxing the execution ensures enterprise-grade security, allowing developers to safely experiment with code-generating agents.

*Massive Scalability*: Shifting the heavy lifting—chunking, embedding, and similarity search—into an isolated cloud environment allows the application to process massive unstructured datasets that would otherwise overwhelm local memory.

*Highly Contextual Retrieval*: Traditional RAG relies on rigid, predefined chunking strategies that often miss context. This agentic approach writes bespoke logic to navigate the data, drastically improving the relevance of the retrieved context.

*Superior Final Output*: By ensuring that the data is meticulously filtered and processed by sub-agents before being handed to a powerful synthesis model, the final generation achieves a higher degree of accuracy and reasoning quality.

### Key Technologies
- **Language**: Go (`go 1.25.0`)
- **SDK**: Google GenAI SDK (`google.golang.org/genai`)
- **Models Used**:
  - `gemini-2.5-flash` (Orchestrator)
  - `gemini-2.5-flash-lite` (Sub-agent worker)
  - `gemini-2.5-pro` (Final Synthesis)
  - `text-embedding-004` (Semantic search / embeddings)
- **Execution Environment**: Vertex AI Reasoning Engines (Cloud Sandbox)

## Architecture Details
1. **Context Generation**: Go dynamically creates a 45MB ultra-massive dataset simulating deep engineering memory leaks and multi-generational medical diagnostic data.
2. **GCS Upload**: The unstructured dataset is uploaded to a Google Cloud Storage bucket (`rag-sandbox-obj-{project}-us-central1`).
3. **Orchestrator Setup**: The Go app spins up an orchestrator with `gemini-2.5-flash`, passing it an `run_rag_agent_logic` tool.
4. **Execution**: The generated Python script is sent to the deployed Vertex AI Reasoning Engine via the `:query` API path attached to a custom executor. It pulls the context from GCS, runs the hybrid RAG logic locally within the container, and returns the top highly-relevant chunks.
5. **Synthesis**: The Orchestrator receives the parsed, high-value chunks and feeds them into `gemini-2.5-pro` for a final, polished reasoning output.

## Prerequisites
- Go 1.25+
- Python 3 & pip
- Google Cloud Project with Vertex AI API enabled
- Authenticated via Application Default Credentials (`gcloud auth application-default login`)

### Environment Variables
You must set the following environment variable before running the application:
- `GOOGLE_CLOUD_PROJECT`: Your Google Cloud Project ID.
- `GOOGLE_CLOUD_LOCATION`: (Optional) Target Vertex AI region. Defaults to `us-central1`.

Note: The application defaults to the **`us-central1`** Vertex AI endpoint, as it is the primary target region for Agent Engine features, but can be overridden.

## Setup & Running

### 1. Provision the Cloud Sandbox
Before running the Go application, you **must** provision the Reasoning Engine and GCS bucket on your Google Cloud Project. This takes about 3-5 minutes.
```bash
export GOOGLE_CLOUD_PROJECT="your-project-id"
./setup_sandbox.sh
```

### 2. Run the RAG Simulation
To run the Go orchestrator:
```bash
export GOOGLE_CLOUD_PROJECT="your-project-id"
go run cmd/sandbox/main.go
```

## Project Structure
- `cmd/sandbox/main.go`: Main application entry point.
- `internal/ai/`: Wrappers and clients for Google GenAI interactions.
- `internal/data/`: Massive dataset generator (1.2M+ lines / 45MB).
- `internal/orchestrator/`: Primary agent orchestration and GenAI loop.
- `internal/python/`: Execution logic for interfacing with Vertex AI Reasoning Engines and local fallback runners.
- `internal/ui/`: Terminal spinners and visual feedback.
