# Agentic RAG Process Workflow (GKE & Gemini 3.1)

This diagram illustrates the multi-model, multi-process lifecycle of a single Agentic RAG request using GKE Sandboxes.

```mermaid
graph TD
    %% Roles and Definitions
    subgraph GoHost [Go Host]
        A[Initialize Orchestrator - 3.1 Flash]
        E[Proxy IPC Request]
        G[Final Synthesis - 3.1 Pro]
    end

    subgraph GKESandbox [GKE Sandbox - gVisor]
        subgraph PythonExecution [Python Worker]
            C[Triage: Regex/Keywords]
            D[Sub-Agent Triage - 3.1 Flash-Lite]
            F[Embed & Cosine Similarity]
            I[Distilled Insight Manifest]
        end
    end

    subgraph GCP_APIs [Google Cloud APIs]
        V[Vertex AI Gemini API]
        M[Vertex AI Embeddings API]
    end

    %% Flow
    A -->|Tiered Discovery| C
    C --> D
    D -->|Vertex AI Call| V
    D --> F
    F -->|Embedding Call| M
    F --> I
    I -->|"IPC: {'type': 'done'}"| G
    G -->|Final Synthesis| V
    G --> H[Final Summary Output]

    %% Styling
    style A fill:#f9f,stroke:#333,stroke-width:2px
    style G fill:#fbb,stroke:#333,stroke-width:2px
    style GKESandbox fill:#eee,stroke:#333,stroke-dasharray: 5 5
    style PythonExecution fill:#fff,stroke:#333
```

## Workflow Steps

1.  **Orchestration**: Go initializes **Gemini 3.1 Flash** with a "Tiered Discovery" strategy.
2.  **Sandbox Provisioning**: Go creates an ephemeral **GKE Job** using the `gvisor` runtime class for secure isolation.
3.  **Triage Phase**: The Python script rapidly scans the mounted dataset using local regex and keyword filters to discard 90% of irrelevant data.
4.  **Sub-Agent Evaluation**: **Gemini 3.1 Flash-Lite** sub-agents evaluate the remaining text blocks to determine their semantic value.
5.  **Vector Search**: Only high-value blocks are sent to the **Vertex AI Embeddings API**. Similarity is calculated locally within the sandbox.
6.  **Insight Manifest**: The script compiles a refined "Insight Manifest" of the most relevant distilled data.
7.  **Final Synthesis**: The manifest is returned to Go, which invokes **Gemini 3.1 Pro** to produce the final polished executive summary.
