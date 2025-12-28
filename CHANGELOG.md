# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2025-12-28

### Added

- **Transparent HTTPS Proxy** - Intercept and log all Claude Code API traffic
- **Activity Monitoring** - Track tool calls, file access, and command execution
- **Policy Enforcement** - Block dangerous operations based on organizational policies
- **Zero Configuration** - Automatic client detection and policy application
- **Multi-Tenant Architecture** - Centralized management across teams and organizations
- **Real-time Log Streaming** - WebSocket-based live activity monitoring
- **Web Dashboard** - Next.js admin UI for monitoring and configuration
- **CLI Tool** - `arfa-cli` for proxy management and authentication

### Security

- JWT-based authentication
- Row-level security for multi-tenancy
- Secure credential storage

[0.1.0]: https://github.com/rastrigin-systems/arfa/releases/tag/v0.1.0
