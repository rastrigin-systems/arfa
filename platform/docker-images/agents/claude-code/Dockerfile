# Claude Code CLI Docker Image
# Based on research: https://docs.claude.com/en/docs/claude-code/setup
#
# This image includes common development tools so Claude Code can:
# - Build Go projects
# - Run Python scripts
# - Use gh CLI for GitHub operations
# - Execute make targets

FROM node:20-bookworm

# Install system dependencies and development tools
RUN apt-get update && \
    apt-get install -y \
        curl \
        git \
        ca-certificates \
        openssh-client \
        bash \
        make \
        build-essential \
        python3 \
        python3-pip \
        python3-venv \
        jq \
        ripgrep \
        fd-find \
    && rm -rf /var/lib/apt/lists/*

# Install Go 1.24
ARG GO_VERSION=1.24.0
RUN curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" | tar -C /usr/local -xzf -
ENV PATH="/usr/local/go/bin:/root/go/bin:$PATH"
ENV GOPATH="/root/go"

# Install gh CLI (GitHub CLI)
RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg \
    && chmod go+r /usr/share/keyrings/githubcli-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | tee /etc/apt/sources.list.d/github-cli.list > /dev/null \
    && apt-get update \
    && apt-get install -y gh \
    && rm -rf /var/lib/apt/lists/*

# Install Claude Code CLI via npm
RUN npm install -g @anthropic-ai/claude-code

# Create necessary directories
RUN mkdir -p /workspace ~/.claude /root/go/pkg/mod

# Copy entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

WORKDIR /workspace

# Claude Code uses stdin/stdout for interaction
ENTRYPOINT ["/entrypoint.sh"]
CMD []
