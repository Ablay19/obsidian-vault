package docs

import (
	"fmt"
	"go/ast"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

type APIGenerator struct {
}

func NewAPIGenerator() *APIGenerator {
	return &APIGenerator{}
}

func (ag *APIGenerator) GenerateAPIDocs(projectRoot string) error {
	// Load packages
	cfg := &packages.Config{
		Dir:  projectRoot,
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedSyntax | packages.NeedDeps,
	}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return err
	}

	apiDir := filepath.Join(projectRoot, "docs", "api")
	if err := os.MkdirAll(apiDir, 0755); err != nil {
		return err
	}

	for _, pkg := range pkgs {
		content := fmt.Sprintf("# Package %s\n\n", pkg.Name)
		content += "## Exported Functions\n\n"
		// Simple, just list functions
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.IsExported() {
					content += fmt.Sprintf("- %s\n", fn.Name.Name)
				}
			}
		}
		filePath := filepath.Join(apiDir, pkg.Name+".md")
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return err
		}
	}
	return nil
}
