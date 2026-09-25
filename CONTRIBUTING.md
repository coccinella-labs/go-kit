# Contributing to Kit

Thank you for your interest in contributing to Kit! This document provides guidelines and instructions for contributing.

## Code of Conduct

This project adheres to a [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## How to Contribute

### Reporting Bugs

- Search existing issues before opening a new one.
- Use the bug report template when opening an issue.
- Include steps to reproduce, expected behavior, and actual behavior.
- Include version information (`kit version`) and OS details.

### Suggesting Features

- Search existing issues before opening a new one.
- Use the feature request template when opening an issue.
- Explain the use case and expected behavior.

### Pull Requests

- Fork the repository and create a feature branch.
- Follow the existing code style.
- Keep changes focused on a single concern.
- Add tests for new functionality.
- Ensure `go vet ./...` and `go test ./...` pass.
- Update documentation if needed, including `make docs`.
- Use the pull request template when opening a PR.

## Development Setup

```bash
git clone https://github.com/coccinella-labs/kit.git
cd kit
make build
make test
make docs
```

See `AGENTS.md` for project conventions used by both human contributors and AI coding agents.

## Project Structure

```
kit/
├── cmd/kit/          CLI entry point
├── internal/         Private application code
│   ├── api/          HTTP client
│   ├── auth/         Authentication
│   ├── commands/     CLI commands
│   ├── config/       Configuration
│   ├── exitcode/     Exit codes
│   └── git/          Git utilities
├── docs/             Generated documentation
├── tools/            Development tools
└── .github/          CI workflows
```

## Commit Guidelines

Use clear, descriptive commit messages. Examples:

- feat: add repository context detection
- fix: handle expired access tokens
- docs: update installation instructions
- test: add integration test for auth
- chore: release v1.0.1

Reference issue numbers when applicable. Keep commits focused on a single change.

## Questions

- Open an issue for general questions.
