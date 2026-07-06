package extension

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func createScriptTemplate(full string) error {
	body := fmt.Sprintf(`#!/usr/bin/env bash
set -euo pipefail

echo "%s extension"
`, full)
	return os.WriteFile(filepath.Join(full, full), []byte(body), 0o755)
}

func createGoTemplate(full string) error {
	moduleName := strings.ReplaceAll(full, "-", "_")
	files := map[string]string{
		"main.go": fmt.Sprintf(`package main

import "fmt"

func main() {
	fmt.Println("%s extension")
}
`, full),
		"go.mod":     fmt.Sprintf("module %s\n\ngo 1.22\n", moduleName),
		".gitignore": fmt.Sprintf("/%[1]s\n/%[1]s.exe\n", full),
		"README.md":  fmt.Sprintf("# %s\n\nBuild locally with:\n\n```sh\ngo build -o %s .\n```\n", full, full),
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(full, name), []byte(body), 0o644); err != nil {
			return err
		}
	}
	return nil
}
