FROM node:18-alpine

WORKDIR /app

# Install Google Gemini CLI
RUN npm install -g @google/gemini-cli

# Set entrypoint to run the gemini-cli
ENTRYPOINT ["gemini"]