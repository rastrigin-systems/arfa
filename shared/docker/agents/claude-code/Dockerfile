# Claude Code CLI Docker Image
# Based on research: https://docs.claude.com/en/docs/claude-code/setup

FROM node:20-bookworm

# Install dependencies
RUN apt-get update && \
    apt-get install -y \
        curl \
        git \
        ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Install Claude Code CLI via npm
RUN npm install -g @anthropic-ai/claude-code

# Create necessary directories
RUN mkdir -p /workspace ~/.claude

# Copy entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

WORKDIR /workspace

# Claude Code uses stdin/stdout for interaction
ENTRYPOINT ["/entrypoint.sh"]
CMD []
