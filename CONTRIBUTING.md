# Contributing to Arfa

Thank you for your interest in contributing to Arfa! This document provides guidelines and information for contributors.

## Quick Links

- [Development Setup](docs/development/getting-started.md)
- [Architecture Overview](docs/architecture/overview.md)
- [Testing Guide](docs/development/testing.md)
- [Code of Conduct](#code-of-conduct)

## Getting Started

1. **Fork the repository** and clone your fork
2. **Set up the development environment** following [Getting Started](docs/development/getting-started.md)
3. **Create a branch** for your changes: `git checkout -b feature/your-feature-name`
4. **Make your changes** with tests
5. **Submit a pull request**

## Development Workflow

### Prerequisites

- Go 1.24+
- PostgreSQL 15+
- Node.js 20+ (for web UI)
- Docker (for local development)

### Running Locally

```bash
# Start dependencies
docker compose up -d

# Install tools and generate code
make install-tools
make generate

# Run tests
make test

# Build all services
make build
```

### Code Generation

We use code generation for type-safe database queries and API types:

```bash
make generate  # Regenerate all code
```

**Never edit files in `generated/`** - they are regenerated from source.

## Pull Request Guidelines

### Before Submitting

- [ ] Run `make test` and ensure all tests pass
- [ ] Run `make lint` and fix any issues
- [ ] Add tests for new functionality
- [ ] Update documentation if needed

### PR Title Format

Use conventional commits format:

- `feat: Add new feature`
- `fix: Fix bug in X`
- `docs: Update documentation`
- `refactor: Refactor X for clarity`
- `test: Add tests for X`
- `chore: Update dependencies`

### Review Process

1. All PRs require at least one approval
2. CI checks must pass
3. Maintain test coverage (target: 85%)

## Architecture

See [Architecture Overview](docs/architecture/overview.md) for details on:

- Service structure (API, CLI, Web)
- Database schema
- Code generation pipeline
- Multi-tenancy design

## Testing

We follow test-driven development (TDD). See [Testing Guide](docs/development/testing.md) for:

- Unit testing patterns
- Integration testing with TestContainers
- Test fixtures and utilities

## Security

### Reporting Vulnerabilities

Please report security vulnerabilities to **security@arfa.dev**. Do not open public issues for security concerns.

### Security Guidelines

- All database queries must include `org_id` for multi-tenancy
- Never commit secrets or credentials
- Use parameterized queries (sqlc handles this)

## Code of Conduct

We are committed to providing a welcoming and inclusive environment. Please:

- Be respectful and constructive
- Welcome newcomers
- Focus on what is best for the community

## Questions?

- Open a [GitHub Discussion](https://github.com/rastrigin-systems/arfa/discussions)
- Check existing [Issues](https://github.com/rastrigin-systems/arfa/issues)

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
