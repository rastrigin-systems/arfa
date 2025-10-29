# MCP Filesystem Server Docker Image
# Official package: @modelcontextprotocol/server-filesystem

FROM node:20-alpine

# Install MCP filesystem server
RUN npm install -g @modelcontextprotocol/server-filesystem

# Copy entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

WORKDIR /workspace
EXPOSE 8001

ENTRYPOINT ["/entrypoint.sh"]
CMD []
