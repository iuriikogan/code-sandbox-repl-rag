# Tiered Routing Test Dataset Design (Ultra-Large Context)

## Purpose
To generate a synthetic dataset for testing the efficiency of agentic RAG against tiered model routing approaches, specifically targeting a context size that exceeds the 1M+ token window of `gemini-1.5-flash` (and `2.5-flash` equivalents). The dataset will focus on deep technical/engineering and medical scenarios, requiring complex, multi-hop reasoning across thousands of disparate chunks of information hidden within massive volumes of noise.

## Approach: Massive Mixed-Domain Interleaved Chunks
The dataset will interleave related facts across an enormous number of chunks. To answer a complex query, the system must perform multi-step retrieval across a context exceeding 1-2 million tokens. This strictly forces the system to rely on the RAG/retrieval mechanism and the reasoning capabilities of the larger tier models, as the entire context cannot be processed in a single prompt by the smaller tier model.

## Data Domains & Scenarios

### 1. Engineering: The "Phantom Memory Leak in Distributed Mesh"
*   **Structure:** Tens of thousands of log lines across 50 microservices, simulated over a 72-hour period.
*   **Scenario:** A slow memory leak in a sidecar proxy (Envoy) only occurs when a specific combination of gRPC headers (from Service Alpha) hits a specific legacy endpoint (on Service Omega) during a daily batch job (Service cron-beta).
*   **Example Spread:**
    *   **Chunk 10,452 (Service Alpha code snippet):** Adds a custom `X-Trace-Legacy-ID` header.
    *   **Chunk 45,901 (Service cron-beta log):** Initiates the daily sync job at 02:00 AM.
    *   **Chunk 89,112 (Envoy Sidecar issue tracker):** "Known issue #442: memory leak when processing non-standard X-Trace headers on endpoint /v1/legacy."
    *   **Chunk 120,555 (Service Omega metrics):** OOM (Out of Memory) kills observed specifically on pods running the `/v1/legacy` endpoint around 02:45 AM.
*   **Target Query:** "Identify the root cause of the OOM kills in Service Omega, including the triggering service and the underlying proxy issue."

### 2. Medical: The "Multi-Generational Genetic Anomaly"
*   **Structure:** Decades of anonymized EHR (Electronic Health Records) for a large family tree (50+ individuals), mixed with thousands of unrelated patient records.
*   **Scenario:** Identifying a rare genetic disorder (e.g., a specific subtype of Porphyria or a mitochondrial disease) that presents differently across generations and genders, exacerbated by a specific environmental trigger (e.g., a common antibiotic prescribed years later).
*   **Example Spread:**
    *   **Chunk 5,200 (Patient A, 1985):** Unexplained severe abdominal pain and photosensitivity.
    *   **Chunk 55,000 (Patient B - Daughter of A, 2010):** Diagnosed with "atypical peripheral neuropathy."
    *   **Chunk 150,222 (Patient C - Son of B, 2025):** Prescribed Sulfonamide antibiotics for a routine infection.
    *   **Chunk 150,900 (Patient C, 2025 - ER visit):** Admitted with acute neurovisceral crisis and dark urine 3 days after starting antibiotics.
    *   **Chunk 200,000 (Medical Journal snippet):** "Sulfonamides are known triggers for acute attacks in Acute Intermittent Porphyria (AIP)."
*   **Target Query:** "Trace the genetic link between Patient A, B, and C, and explain the acute ER admission of Patient C."

### 3. The "Ocean of Noise" (Context Bloat)
To exceed the 1M+ token limit (approx. 3-4MB of raw text, we will generate ~20-50MB of text):
*   **System Logs:** Millions of lines of normal HTTP 200 requests, garbage collection pauses, and routine background syncs.
*   **Medical Records:** Thousands of normal blood panels, routine physicals, and common cold diagnoses.
*   **Filler Text:** Randomly generated corporate policy documents, generic code comments, and standard operating procedures to pad the token count.

## Implementation Details
1.  **Massive Generator (`internal/data/generator.go`):**
    *   Create a scalable generation function that can produce tens of megabytes of text.
    *   Use a seeded random number generator to ensure reproducibility.
    *   Inject the "scenario needles" at predefined, widely separated indices.
2.  **Dataset File (`internal/data/data.go`):**
    *   Update `GenerateMassiveContext` to stream generation directly to disk (to avoid high memory usage during generation) until the target file size (e.g., 50MB) is reached.
3.  **Test Queries:**
    *   Ensure queries cannot be answered by simple keyword matching but require understanding the relationships between the far-flung chunks.

## Success Criteria
*   The generated file size exceeds the target model's token limit (e.g., > 1.5M tokens / ~6MB of dense text).
*   The scenarios demand multi-step, relational retrieval (A points to B, B points to C).
*   A single-pass prompt with the entire file fails (due to context limits), proving the necessity of the RAG/Routing system.
