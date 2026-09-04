<p align="center">
  <img src="https://raw.githubusercontent.com/bniladridas/kit/main/.github/assets/thumbnail.png" alt="kit" width="100%">
</p>

# Kit

A lightweight developer CLI built with Go for working with Git hosting platforms from the terminal.

## Features

- GitHub authentication with OAuth device flow or personal access tokens
- Repository, issue, and pull request workflows
- Automatic repository context detection
- Shell completion for Bash, Zsh, and Fish
- Structured output for scripting

## Installation

### Homebrew

```bash
brew install bniladridas/kit/kit
```

For Homebrew installation, see:
https://github.com/bniladridas/homebrew-kit

### Install Script

For Linux and macOS:

```bash
curl -fsSL https://raw.githubusercontent.com/bniladridas/kit/main/scripts/install.sh | sh
```

### Go

```bash
go install github.com/bniladridas/kit/cmd/kit@latest
```

### Manual Download

Download the binary from the [latest release](https://github.com/bniladridas/kit/releases/latest) and move it to your PATH.

```bash
curl -fsSL https://github.com/bniladridas/kit/releases/download/v1.0.0/kit-linux-amd64 -o kit
chmod +x kit
sudo mv kit /usr/local/bin
```

### Update

```bash
kit update
```

Checks for a newer version and suggests the correct update command. It does not update automatically.

| Installed with | Update command                                     |
|----------------|---------------------------------------------------|
| Homebrew       | `brew upgrade kit`                                |
| Go             | `go install github.com/bniladridas/kit/cmd/kit@latest` |
| Install script | Run the install script again                      |
| Manual download| Download the latest release and replace the binary |

## Quick Start

```bash
kit auth login github
kit github repo list
kit github issue list
kit github pr list
```

Or with a personal access token:

```bash
kit auth login github --token <token>
```

## Commands

### Authentication

```bash
kit auth login github
kit auth login github --token <token>
kit auth status
kit auth whoami
kit auth logout
```

- `kit auth login github` - authenticate via OAuth device flow
- `kit auth login github --token <token>` - authenticate with a personal access token
- `kit auth status` - show authentication status for all providers
- `kit auth whoami` - show the current authenticated user
- `kit auth logout` - remove stored credentials

### GitHub Integration

```bash
kit github repo list
kit github repo list --json
kit github repo clone owner/repo
kit github issue list
kit github issue create --title "Bug" --body "Description"
kit github pr list
kit github pr create --title "Fix"
kit github pr checkout 123
```

- `kit github repo list [--json] [--quiet]` - list repositories
- `kit github repo clone <owner/repo>` - clone a repository
- `kit github issue list [--owner --repo] [--state] [--json] [--quiet]` - list issues
- `kit github issue create --title <title> [--owner --repo] [--body]` - create an issue
- `kit github pr list [--owner --repo] [--state] [--json] [--quiet]` - list pull requests
- `kit github pr create --title <title> [--owner --repo] [--body]` - create a pull request
- `kit github pr checkout <number> [--owner --repo]` - checkout a pull request branch

Owner and repo are inferred from the current Git repository when not provided.

### Configuration

```bash
kit config list
kit config set <key> <value>
kit config get <key>
```

### Context

```bash
kit context
```

Show the current Git repository context: owner, repo, branch, and remote.

### Shell Completion

```bash
kit completion bash
kit completion zsh
kit completion fish
```

### Diagnostics

```bash
kit doctor
```

Check configuration, authentication, Git availability, and network connectivity.

### Version

```bash
kit version
```

Example:

```text
kit version 1.0.0
```

## Compatibility Policy

- Stable commands: auth, github, config, context, completion, doctor, version
- Command names, flags, and output formats will not change in patch releases.
- Breaking changes will only occur in major versions with advance notice.
- Features will be deprecated in one minor release before removal in the next major release.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid usage |
| 3 | Not authenticated |
| 4 | Configuration error |
| 5 | Network error |
| 6 | Git error |
| 7 | API error |
| 8 | Cancelled |
| 9 | Not implemented |

## Environment Variables

### Runtime configuration

- `KIT_API_URL` - override the GitHub API URL
- `KIT_CONFIG` - path to config file

### Development

- `KIT_GITHUB_TOKEN` - GitHub personal access token for live integration tests
- `KIT_INTEGRATION_TESTS` - set to `1` to enable live integration tests

## Development

```bash
git clone https://github.com/bniladridas/kit.git
cd kit
make build
make test
make docs
```

## License

MIT
