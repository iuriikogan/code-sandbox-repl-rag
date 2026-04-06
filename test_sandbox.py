import logging
import httplib2
httplib2.debuglevel = 4

import requests
import logging
import http.client
http.client.HTTPConnection.debuglevel = 1

logging.basicConfig()
logging.getLogger().setLevel(logging.DEBUG)
requests_log = logging.getLogger("requests.packages.urllib3")
requests_log.setLevel(logging.DEBUG)
requests_log.propagate = True

from google.cloud import aiplatform

aiplatform.init(project="develop-491110", location="us-central1")
agent_engine = aiplatform.reasoning_engines.ReasoningEngine()
sandbox = agent_engine.create_sandbox(code_language="PYTHON")
print("Sandbox:", sandbox.sandbox_id)
