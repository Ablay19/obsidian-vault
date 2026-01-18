package docs

import (
	"os"
	"path/filepath"
)

type SetupGenerator struct {
}

func NewSetupGenerator() *SetupGenerator {
	return &SetupGenerator{}
}

func (sg *SetupGenerator) GenerateSetupGuide(projectRoot string) error {
	content := `# Development Setup Guide

## Prerequisites

- Go 1.25.4 or later
- Git

## Setup Steps

1. Clone the repository:
   ` + "```bash" + `
   git clone <repository-url>
   cd obsidian-vault
   ` + "```" + `

2. Install dependencies:
   ` + "```bash" + `
   go mod download
   ` + "```" + `

3. Build the CLI:
   ` + "```bash" + `
   go build -o mauritania-cli ./cmd/mauritania-cli
   ` + "```" + `

4. Run tests:
   ` + "```bash" + `
   go test ./...
   ` + "```" + `

## Development

- Use ` + "`go run ./cmd/mauritania-cli`" + ` for development
- Run ` + "`go vet ./...`" + ` for linting
- Use ` + "`go test -cover ./...`" + ` for coverage

## Contributing

1. Create a feature branch
2. Make changes
3. Run tests and linting
4. Submit a pull request
`
	filePath := filepath.Join(projectRoot, "docs", "development", "setup.md")
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}
	return os.WriteFile(filePath, []byte(content), 0644)
}
