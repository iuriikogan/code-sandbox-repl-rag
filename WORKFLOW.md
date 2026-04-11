# Agentic RAG Process Workflow

This diagram illustrates the multi-model, multi-process lifecycle of a single Agentic RAG request using a Local IPC Swarm architecture and Pure-Python analytical algorithms.

```mermaid
graph TD
    %% Roles and Definitions
    subgraph GoHost [Go Host Process]
        A[Initialize Orchestrator]
        B[Generate 45MB+ Noise Context]
        G[Final Synthesis Step]
        L[IPC Swarm Router]
    end

    subgraph Orchestrator [Orchestrator Agent: Gemini 3.1 Flash Lite]
        C[Generate Hybrid RAG Python Script]
    end

    subgraph LocalFileSystem [Local File System]
        H[(Context Dataset File)]
    end

    subgraph PythonExecution [Python Subprocess REPL]
        D[Read Context File]
        E[Lexical Filter: BM25/TF-IDF]
        F[Semantic Search: IPC Batch Embeddings]
        I[Dynamic Relevance Filter: Std Dev Outliers]
        J[Recursive Vector Expansion: Rocchio Algorithm]
    end

    %% Flow
    A -->|System Instructions| C
    B -.->|Writes File| H
    
    C -->|run_rag_agent_logic| D
    D --> E
    E --> F
    F --> I
    I -->|Swarm Output Clues| J
    J -.->|Recursive Query Update| E
    
    H -.->|File Stream| D
    
    I <-->|ipc_batch_call| L
    F <-->|ipc_batch_embed| L
    L <-->|Concurrent API Batches| M((Gemini 2.5 Flash / Embeddings API))
    
    J -->|submit_clues_for_synthesis| G
    G -.->|Vertex AI / AI Studio: gemini-2.5-pro| G
    G --> K[Final Summary Output]

    %% Styling
    style A fill:#f9f,stroke:#333,stroke-width:2px
    style C fill:#bbf,stroke:#333,stroke-width:2px
    style PythonExecution fill:#eee,stroke:#333,stroke-dasharray: 5 5
    style G fill:#fbb,stroke:#333,stroke-width:2px
    style L fill:#bff,stroke:#333,stroke-width:2px
```

## Workflow Steps

1.  **Context Generation**: Go initializes an ultra-massive unstructured dataset filled with distractor noise and isolated "needle" strings.
2.  **File Storage**: The dataset is written to a local temp file, bypassing standard HTTP payload limits.
3.  **Orchestration**: Go initializes `gemini-3.1-flash-lite-preview` with a set of "cost-optimizing" analytical search instructions.
4.  **Code Generation**: The Orchestrator generates a specialized Python script tailored to the dynamic query, utilizing pre-injected analytical tools like `BM25`.
5.  **Execution Environment**: The Go host invokes a local Python 3 subprocess.
6.  **Lexical Filter (Python)**: Python parses millions of lines and aggressively shrinks them down to ~500 top candidates natively using a BM25 TF-IDF approximation.
7.  **Embedding Generation (IPC)**: Python obtains vectors for the 500 candidate chunks by batching them over an IPC (Standard I/O) JSON pipe to the Go host, which calls `text-embedding-004` (or `gemini-embedding-001`).
8.  **Vector Search (Python)**: Cosine Similarity and Standard Deviation calculation are run locally in Python to avoid high LLM context costs. Outlier chunks are selected.
9.  **Swarm Map-Reduce (IPC)**: The highest-scoring chunks are dispatched to a concurrent swarm of `gemini-2.5-flash` Sub-Agents. These sub-agents run simultaneously to extract concise causal evidence and text clues from the chunks.
10. **Recursive Vector Expansion**: Extracted clues are re-embedded. The query vector is mathematically updated (`update_vector_rocchio`), dragging the semantic center toward the new clues. The script recursively repeats steps 6-9, tracing multi-hop context across the dataset.
11. **Final Synthesis**: Only the highly compressed "distilled" clues are returned to Go and sent to `gemini-2.5-pro` for a polished, highly accurate reasoning output.