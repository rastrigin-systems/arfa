# MCP Git Server Docker Image
# Official package: mcp-server-git (Python-based)
# Note: Official MCP git server is Python, not Node.js

FROM python:3.11-slim

# Install git and dependencies
RUN apt-get update && \
    apt-get install -y git && \
    rm -rf /var/lib/apt/lists/*

# Install MCP git server via pip
RUN pip install --no-cache-dir mcp-server-git

# Copy entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

WORKDIR /workspace
EXPOSE 8002

ENTRYPOINT ["/entrypoint.sh"]
CMD []
