# Agentic RAG Process Workflow

This diagram illustrates the multi-model, multi-process lifecycle of a single Agentic RAG request.

```mermaid
graph TD
    %% Roles and Definitions
    subgraph GoHost [Go Host]
        A[Initialize Orchestrator]
        E[Proxy IPC Request]
        G[Final Synthesis Step]
    end

    subgraph Orchestrator [Orchestrator Agent - Flash]
        B[Generate Python Script]
    end

    subgraph PythonExecution [Python Execution Environment]
        C[Chunk Dataset]
        D[Calculate Cosine Similarity]
        F[Dynamic Relevance Filter]
        I[Vertex AI SDK Direct Call]
    end

    subgraph VertexSandbox [Vertex AI Agent Engine Sandbox]
        J[Secure Isolated Environment]
    end

    %% Flow
    A -->|System Instructions| B
    B -->|execute_python_script| C
    
    %% Local Path
    C -.->|"IPC (LocalRunner)"| E
    E -.->|Vertex AI API| E
    E -.->|"Vector Result"| D

    %% Sandbox Path
    C -.->|"Direct SDK (SandboxRunner)"| I
    I -.->|Vertex AI API| I
    I -.->|"Vector Result"| D
    
    C --- J
    D --- J
    
    D --> F
    F -->|"IPC: {'type': 'done'}"| G
    G -.->|Vertex AI: gemini-3.1-pro-preview| G
    G --> H[Final Summary Output]

    %% Styling
    style A fill:#f9f,stroke:#333,stroke-width:2px
    style B fill:#bbf,stroke:#333,stroke-width:2px
    style C fill:#bfb,stroke:#333,stroke-width:2px
    style G fill:#fbb,stroke:#333,stroke-width:2px
    style J fill:#eee,stroke:#333,stroke-dasharray: 5 5
```

## Workflow Steps

1.  **Orchestration**: Go initializes Gemini Flash with a set of "cost-optimizing" instructions.
2.  **Code Generation**: Flash generates a specialized Python script for the specific query.
3.  **Execution Environment**:
    - **LocalRunner**: Executes Python via `os/exec`. Communicates with Go via a JSON IPC channel for embeddings and sub-agent calls.
    - **SandboxRunner**: Executes Python in a secure Vertex AI Agent Engine Sandbox. Python can use the `vertexai` SDK directly for embeddings and sub-agents, or continue to use IPC if desired.
4.  **Data Processing**: Python reads the context file (`context.txt` in sandbox or local) and chunks it.
5.  **Embedding Generation**: Python obtains vectors for each chunk either via IPC (Go proxy) or directly via the Vertex AI SDK.
6.  **Vector Search**: Similarity is calculated locally in Python to avoid high LLM context costs.
7.  **Dynamic Filtering**: The script dynamically selects the most relevant content (e.g. > 0.75 similarity).
8.  **Final Synthesis**: Only the "distilled" chunks are sent to Gemini Pro for the final high-quality summary.
