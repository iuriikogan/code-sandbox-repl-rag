import os
os.environ["GOOGLE_API_USE_CLIENT_CERTIFICATE"] = "false"
os.environ["GOOGLE_API_USE_MTLS"] = "false"

import vertexai
from vertexai.preview.reasoning_engines import ReasoningEngine

project_id = os.environ.get("GOOGLE_CLOUD_PROJECT")
location = "us-central1"
bucket_name = f"rag-sandbox-obj-{project_id}-{location}"

vertexai.init(project=project_id, location=location, staging_bucket=f"gs://{bucket_name}")

class CloudExecutor:
    def set_up(self):
        from google.cloud import storage
        self.client = storage.Client()

    def query(self, code: str, gcs_context_uri: str) -> str:
        import sys
        import io
        import traceback
        import os
        
        try:
            bucket_name = gcs_context_uri.split('/')[2]
            blob_name = '/'.join(gcs_context_uri.split('/')[3:])
            bucket = self.client.bucket(bucket_name)
            blob = bucket.blob(blob_name)
            blob.download_to_filename("context.txt")
        except Exception as e:
            return f"Failed to download context from GCS: {e}"
            
        old_stdout = sys.stdout
        new_stdout = io.StringIO()
        sys.stdout = new_stdout
        try:
            exec(code, globals())
            return new_stdout.getvalue()
        except Exception:
            return f"Execution Error:\n{traceback.format_exc()}\nOutput so far:\n{new_stdout.getvalue()}"
        finally:
            sys.stdout = old_stdout

print("Creating Reasoning Engine... This will take a few minutes as Vertex AI builds the container.")
engine = ReasoningEngine.create(
    CloudExecutor(),
    display_name="rag-simulation-engine",
    requirements=["google-cloud-aiplatform", "google-cloud-storage"]
)
print(f"Deployment complete! Engine ID: {engine.resource_name}")
