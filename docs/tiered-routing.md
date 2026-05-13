# Tiered Routing Architecture

## Overview

The Tiered Routing Architecture is designed to optimize the performance, cost, and accuracy of Agentic RAG (Retrieval-Augmented Generation) systems when dealing with massive context windows (1M+ tokens). By intelligently routing queries and tasks across a hierarchy of Gemini models, the system ensures that high-complexity reasoning is handled by the most capable models, while simpler tasks are executed efficiently by faster, lower-cost models.

## Model Hierarchy

We utilize three tiers of Gemini models to balance efficiency and reasoning depth:

### Tier 1: Gemini 3.1 Flash-Lite
*   **Role:** Fast Triage & Simple Retrieval.
*   **Characteristics:** Lowest latency and cost.
*   **Use Cases:**
    *   Initial query intent classification.
    *   Simple keyword-based lookups.
    *   Metadata extraction from small chunks.
    *   Formatting and cleaning of retrieved results.

### Tier 2: Gemini 3.1 Flash
*   **Role:** Intermediate Reasoning & Multi-Hop Retrieval.
*   **Characteristics:** Balanced performance with significant reasoning capabilities.
*   **Use Cases:**
    *   Multi-step reasoning across related chunks (e.g., correlating logs from two services).
    *   Summarization of medium-sized context segments (up to 128k tokens).
    *   Domain-specific analysis (Engineering/Medical) that requires understanding of technical terminology.

### Tier 3: Gemini 3.1 Pro
*   **Role:** Complex Reasoning & Ultra-Large Context Analysis.
*   **Characteristics:** High-fidelity reasoning and support for 1M+ token context windows.
*   **Use Cases:**
    *   Deep multi-hop reasoning across thousands of disparate chunks hidden in noise.
    *   Final synthesis of complex scenarios like the "Phantom Memory Leak" or "Multi-Generational Genetic Anomaly."
    *   Processing the entire "Ocean of Noise" dataset when specific needles cannot be found via lower-tier retrieval.

## Routing Logic

The system employs an Orchestrator that follows a multi-stage routing process:

1.  **Complexity Estimation:** The Orchestrator analyzes the incoming query to determine if it requires simple fact retrieval or complex, relational reasoning.
2.  **Context Scoping:** Based on the query, the system estimates the volume of context required. If the context exceeds the limits of Tier 1 or 2, it is automatically escalated.
3.  **Tier Selection:**
    *   **Direct Path:** Simple queries are routed directly to Tier 1.
    *   **Escalation Path:** If Tier 1 fails to find a high-confidence answer, or if the query is identified as complex, it is escalated to Tier 2.
    *   **Pro Path:** Queries requiring global context understanding or extreme multi-hop reasoning across massive data volumes are routed to Tier 3.
4.  **Verification & Feedback:** Higher tiers can verify the outputs of lower tiers, providing a feedback loop to improve routing accuracy over time.

## Documentation & Memory Routing

In addition to query routing, the project adheres to a tiered strategy for managing persistent context, instructions, and memory. Every fact or preference must be routed to exactly one tier:

### Tier 1: Project-Wide Instructions (`GEMINI.md`)
*   **Scope:** Team-shared conventions, architecture rules, and repo-wide workflows.
*   **Location:** Root or major directory level.
*   **Retention:** Committed to the repository and shared with all contributors.

### Tier 2: Scoped Instructions (`sub-dir/GEMINI.md`)
*   **Scope:** Highly specific overrides or guidelines for a particular subdirectory.
*   **Location:** Within the relevant subdirectory.
*   **Retention:** Committed to the repository; supersedes Tier 1 for its scope.

### Tier 3: Private Project Memory (`MEMORY.md`)
*   **Scope:** Personal-to-the-user local setup, machine-specific notes, or private workflows.
*   **Location:** Local private memory folder (e.g., `~/.gemini/tmp/docs/memory/`).
*   **Retention:** **Never** committed to the repo. `MEMORY.md` serves as an index for sibling markdown notes.

### Tier 4: Global Personal Memory (`~/.gemini/GEMINI.md`)
*   **Scope:** Cross-project personal preferences and durable facts (e.g., "I prefer Go over Python").
*   **Location:** Global user configuration directory.
*   **Retention:** Follows the user across all workspaces; never workspace-specific.

## Data Domains & Scenarios

The tiered routing is specifically tested against complex domains:

### Engineering: Distributed Mesh Diagnostics
Requires correlating thousands of log lines and service metrics across disparate systems (e.g., Service Alpha headers vs. Service Omega OOM kills).

### Medical: Genetic Anomaly Tracking
Requires tracing genetic links and environmental triggers across decades of EHR data and family trees, identifying rare disorders (e.g., Porphyria subtypes).

## Success Metrics
*   **Routing Accuracy:** Percentage of queries correctly routed to the lowest viable tier.
*   **Latency Optimization:** Reduction in end-to-end response time compared to a single-tier (Pro-only) approach.
*   **Cost Efficiency:** Reduction in total token cost through strategic model usage.
*   **Reasoning Fidelity:** Ability to solve complex multi-hop scenarios within the "Ocean of Noise."
