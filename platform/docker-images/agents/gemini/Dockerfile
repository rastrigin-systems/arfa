FROM node:22-alpine

WORKDIR /app

# Install build dependencies for native modules (python3, make, g++)
RUN apk add --no-cache python3 make g++

# Install Google Gemini CLI
RUN npm install -g @google/gemini-cli

# Set entrypoint to run the gemini-cli
ENTRYPOINT ["gemini"]