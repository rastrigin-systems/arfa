FROM python:3.11-slim

WORKDIR /app

# Install Google GenAI SDK
RUN pip install google-genai

# Copy entrypoint script
COPY entrypoint.py /app/entrypoint.py

# Set entrypoint
ENTRYPOINT ["python", "/app/entrypoint.py"]
