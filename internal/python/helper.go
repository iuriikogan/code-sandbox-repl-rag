package python

// HelperCode contains the Python helper library injected into the agent's scripts.
const HelperCode = `
import json
import sys
import os
import math

class RAG:
    """Helper class for RAG operations, abstracting IPC and Cloud SDKs."""
    
    _is_sandbox_cached = None

    @staticmethod
    def _is_sandbox():
        if RAG._is_sandbox_cached is None:
            RAG._is_sandbox_cached = os.environ.get('PROJECT_ID') is not None
        return RAG._is_sandbox_cached

    @staticmethod
    def get_embedding(text):
        """Fetches an embedding vector for the given text."""
        if RAG._is_sandbox():
            from vertexai.language_models import TextEmbeddingModel
            model = TextEmbeddingModel.from_pretrained("text-embedding-004")
            embeddings = model.get_embeddings([text])
            return embeddings[0].values
        else:
            print(json.dumps({"type": "embed", "chunk": text}))
            sys.stdout.flush()
            line = sys.stdin.readline()
            if not line:
                return []
            return json.loads(line).get("vector", [])

    @staticmethod
    def run_sub_agent(instruction, chunk):
        """Dispatches a task to a sub-agent (Gemini Flash)."""
        if RAG._is_sandbox():
            from vertexai.generative_models import GenerativeModel
            worker = GenerativeModel("gemini-2.5-flash")
            response = worker.generate_content(f"Instruction: {instruction}\nData: {chunk}")
            return response.text
        else:
            print(json.dumps({"type": "call", "instruction": instruction, "chunk": chunk}))
            sys.stdout.flush()
            line = sys.stdin.readline()
            if not line:
                return "Error: No response from Go host"
            return json.loads(line).get("result", "")

    @staticmethod
    def cosine_similarity(v1, v2):
        """Calculates cosine similarity between two vectors."""
        if not v1 or not v2:
            return 0.0
        dot_product = sum(a * b for a, b in zip(v1, v2))
        mag1 = math.sqrt(sum(a * a for a in v1))
        mag2 = math.sqrt(sum(a * a for a in v2))
        if not mag1 or not mag2:
            return 0.0
        return dot_product / (mag1 * mag2)

    @staticmethod
    def finish(output):
        """Returns the final results and terminates the script."""
        if RAG._is_sandbox():
            # In sandbox, we just print the final output; the runner captures it.
            print("--- FINAL OUTPUT ---")
            print(output)
        else:
            print(json.dumps({"type": "done", "output": output}))
            sys.stdout.flush()
        sys.exit(0)

    @staticmethod
    def get_context_path():
        """Returns the path to the context file."""
        return os.environ.get("CONTEXT_FILE", "context.txt")

# Initialize RAG helper
rag = RAG()
`
