# Agentic RAG Process Workflow

This diagram illustrates the multi-model, multi-process lifecycle of a single Agentic RAG request.

```mermaid
graph TD
    %% Roles and Definitions
    subgraph GoHost [Go Host]
        A[Initialize Orchestrator]
        E[Proxy Embedding Req]
        G[Final Synthesis Step]
    end

    subgraph Orchestrator [Orchestrator Agent - Flash]
        B[Generate Python Script]
    end

    subgraph PythonSandbox [Python Sandbox - os/exec]
        C[Chunk Dataset]
        D[Calculate Cosine Similarity]
        F[Dynamic Relevance Filter]
    end

    %% Flow
    A -->|System Instructions| B
    B -->|execute_python_script| C
    C -->|IPC: {'type': 'embed'}| E
    E -.->|Vertex AI: text-embedding-004| E
    E -->|IPC: {'vector': [... ]}| D
    D --> F
    F -->|IPC: {'type': 'done'}| G
    G -.->|Vertex AI: gemini-3.1-pro| G
    G --> H[Final Summary Output]

    %% Styling
    style A fill:#f9f,stroke:#333,stroke-width:2px
    style B fill:#bbf,stroke:#333,stroke-width:2px
    style C fill:#bfb,stroke:#333,stroke-width:2px
    style G fill:#fbb,stroke:#333,stroke-width:2px
```

## Workflow Steps

1.  **Orchestration**: Go initializes Gemini Flash with a set of "cost-optimizing" instructions.
2.  **Code Generation**: Flash generates a specialized Python script for the specific query.
3.  **Local Processing**: Python reads the massive context file locally and chunks it.
4.  **IPC Embedding**: Python requests vectors for each chunk via a JSON-based IPC channel with the Go host.
5.  **Vector Search**: Similarity is calculated locally in Python to avoid high LLM context costs.
6.  **Dynamic Filtering**: The script dynamically selects the most relevant content (e.g. > 0.75 similarity).
7.  **Final Synthesis**: Only the "distilled" chunks are sent to Gemini Pro for the final high-quality summary.
