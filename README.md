# Aigile

Aigile is a CLI tool that helps you generate Epics, Features, User Stories, and Tasks using LLMs (OpenAI, Gemini, Azure OpenAI) and integrates with GitHub Projects or Azure DevOps.

## Features

- Generate agile items in multiple languages (English by default)
- Read items from XLSX files with structured input
- Generate content using different LLM providers:
  - OpenAI (default)
  - Gemini (coming soon)
  - Azure OpenAI (coming soon)
- Integration with project management tools:
  - GitHub Projects
  - Azure DevOps (coming soon)

## Installation

```bash
go install github.com/leocomelli/aigile@latest
```

## Configuration

The following environment variables are required:

### LLM Configuration
- `LLM_PROVIDER`: The LLM provider to use (default: "openai")
- `LLM_API_KEY`: Your LLM API key
- `LLM_MODEL`: The model to use (e.g., "gpt-4" for OpenAI)
- `LLM_ENDPOINT`: The API endpoint (required for Azure OpenAI)

### GitHub Configuration
- `GITHUB_TOKEN`: Your GitHub personal access token
- `GITHUB_OWNER`: The owner of the repository
- `GITHUB_REPO`: The repository name

## Usage

1. Prepare your XLSX file with the following columns:
   - Type: The type of item (Epic, Feature, User Story, Task)
   - Parent: ID/reference of the parent item (e.g., Epic ID for Features)
   - Context: Description of what needs to be generated
   - Criteria: Additional validation criteria (optional)

2. Run the generate command:
```bash
# Generate items in English (default)
aigile generate -f path/to/your/file.xlsx

# Generate items in a different language
aigile generate -f path/to/your/file.xlsx -l portuguese
```

3. Additional flags:
```bash
# Set log level
aigile generate -f file.xlsx --log-level debug

# Show help
aigile generate --help
```

## Development

### Prerequisites

- Go 1.22 or higher
- Make

### Common Tasks

```bash
# Install dependencies
make deps

# Build the binary
make build

# Run tests
make test

# Run integration tests
make integration-test

# Run linter
make lint

# Clean build artifacts
make clean

# Update dependencies
make tidy
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details. 