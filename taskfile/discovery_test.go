package taskfile

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write %s: %v", name, err)
	}
	return path
}

func TestDiscoveryPrefersYAMLOverHCL(t *testing.T) {
	dir := t.TempDir()
	yamlPath := writeFile(t, dir, "Taskfile.yml", "version: '3'\n")
	writeFile(t, dir, "Taskfile.hcl", "version = 3\n")

	node, err := NewFileNode("", dir)
	if err != nil {
		t.Fatalf("NewFileNode returned error: %v", err)
	}
	if node.Location() != yamlPath {
		t.Fatalf("expected %s, got %s", yamlPath, node.Location())
	}
}

func TestHCLLoaderInvoked(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "Taskfile.hcl", "version = \"3\"\n")

	node, err := NewFileNode("", dir)
	if err != nil {
		t.Fatalf("NewFileNode returned error: %v", err)
	}
	r := NewReader()
	_, err = r.Read(context.Background(), node)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExtensionlessTaskfile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "Taskfile", "version = \"3\"\n")

	node, err := NewFileNode("", dir)
	if err != nil {
		t.Fatalf("NewFileNode returned error: %v", err)
	}
	r := NewReader()
	_, err = r.Read(context.Background(), node)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
