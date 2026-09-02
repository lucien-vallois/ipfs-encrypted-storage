# Contributing to IPFS Encrypted Storage

Thank you for your interest in contributing to IPFS Encrypted Storage! This document provides guidelines and information for contributors.

## Development Setup

### Prerequisites

- Go 1.21 or later
- IPFS daemon (for testing)
- Git

### Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/your-username/ipfs-encrypted-storage.git
   cd ipfs-encrypted-storage
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Run tests:
   ```bash
   go test ./...
   ```

4. Build the project:
   ```bash
   go build -o ipfs-storage ./src
   ```

## Development Workflow

### 1. Choose an Issue

- Check the [Issues](https://github.com/your-username/ipfs-encrypted-storage/issues) page for open tasks
- Comment on the issue to indicate you're working on it
- Create a new branch for your work

### 2. Create a Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/issue-number-description
```

### 3. Make Changes

- Write clear, concise commit messages
- Follow the existing code style
- Add tests for new functionality
- Update documentation as needed

### 4. Test Your Changes

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run benchmarks
go test -bench=. ./tests

# Run specific tests
go test -run TestSpecificFunction ./tests
```

### 5. Submit a Pull Request

- Ensure all tests pass
- Update CHANGELOG.md if needed
- Write a clear PR description
- Reference any related issues

## Code Guidelines

### Go Style

- Follow standard Go formatting (`go fmt`)
- Use `gofmt` for consistent formatting
- Write meaningful variable and function names
- Add comments for exported functions and types

### Testing

- Write unit tests for all new functions
- Aim for good test coverage
- Use table-driven tests where appropriate
- Test error conditions and edge cases

### Documentation

- Update README.md for significant changes
- Add examples for new features
- Keep code comments up to date

## Security Considerations

This project handles cryptographic operations and sensitive data. Please consider:

- Never commit sensitive information
- Use secure random number generation
- Follow cryptographic best practices
- Validate all inputs
- Handle errors securely

## Reporting Issues

- Use the [Issues](https://github.com/your-username/ipfs-encrypted-storage/issues) page
- Provide clear steps to reproduce
- Include relevant system information
- Be respectful and constructive

## Code of Conduct

This project follows a code of conduct to ensure a welcoming environment for all contributors. By participating, you agree to:

- Be respectful and inclusive
- Focus on constructive feedback
- Accept responsibility for mistakes
- Show empathy towards other contributors
- Help create a positive community

## License

By contributing to this project, you agree that your contributions will be licensed under the same license as the project (MIT License).
