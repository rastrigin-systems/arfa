# Gemini CLI Docker Image
#
# This image includes common development tools so Gemini can:
# - Build Go projects
# - Run Python scripts
# - Use gh CLI for GitHub operations
# - Execute make targets

FROM node:22-bookworm

WORKDIR /workspace

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

# Install Google Gemini CLI
RUN npm install -g @google/gemini-cli

# Create necessary directories
RUN mkdir -p /workspace /root/go/pkg/mod

# Set entrypoint to run the gemini-cli
ENTRYPOINT ["gemini"]
