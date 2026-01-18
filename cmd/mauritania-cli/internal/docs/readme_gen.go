package docs

import (
	"os"
	"path/filepath"
)

type ReadmeGenerator struct {
}

func NewReadmeGenerator() *ReadmeGenerator {
	return &ReadmeGenerator{}
}

func (rg *ReadmeGenerator) GenerateReadme(projectRoot string) error {
	content := `# Mauritania CLI

A command-line interface that enables remote development and project management through Mauritanian network provider services.

## Installation

` + "```bash" + `
go build -o mauritania-cli ./cmd/mauritania-cli
` + "```" + `

## Usage

` + "```bash" + `
./mauritania-cli --help
` + "```" + `

## Features

- Remote development support
- Project management
- Network services integration

## Documentation

See [docs/](docs/) for comprehensive documentation.

## License

MIT
`
	filePath := filepath.Join(projectRoot, "README.md")
	return os.WriteFile(filePath, []byte(content), 0644)
}
