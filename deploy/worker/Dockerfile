FROM python:3.11-slim

WORKDIR /app

# Install dependencies
RUN pip install --no-cache-dir \
    google-cloud-aiplatform \
    google-genai \
    pandas \
    scikit-learn \
    numpy

# The script is now mounted via ConfigMap at /mnt/code/script.py
CMD ["python"]
