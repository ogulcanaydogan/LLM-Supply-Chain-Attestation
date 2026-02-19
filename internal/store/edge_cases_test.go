package store

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// --- SaveLocal: destination file create error ---

func TestSaveLocal_DestinationCreateError(t *testing.T) {
	srcDir := t.TempDir()
	srcFile := filepath.Join(srcDir, "src.json")
	os.WriteFile(srcFile, []byte("data"), 0o644)

	// Destination is a file, not a directory — cannot create a child file inside it
	dstFile := filepath.Join(t.TempDir(), "not-a-dir")
	os.WriteFile(dstFile, []byte("blocker"), 0o644)

	_, err := SaveLocal(srcFile, filepath.Join(dstFile, "subdir"))
	if err == nil {
		t.Fatal("expected error when destination parent is a file")
	}
}

// --- PullOCI: write to invalid output path ---

func TestPullOCI_WriteError(t *testing.T) {
	host := startRegistry(t)

	// First publish a bundle
	dir := t.TempDir()
	bundlePath := filepath.Join(dir, "bundle.json")
	os.WriteFile(bundlePath, []byte(`{"data":"test"}`), 0o644)

	ref := fmt.Sprintf("%s/test/write-error:v1", host)
	if _, err := PublishOCI(bundlePath, ref); err != nil {
		t.Fatalf("PublishOCI setup: %v", err)
	}

	// Pull to a path inside a non-existent deeply nested directory
	err := PullOCI(ref, "/nonexistent/deep/path/out.json")
	if err == nil {
		t.Fatal("expected error when output path is non-writable")
	}
}

// --- PublishOCI: registry connection refused ---

func TestPublishOCI_RegistryConnectionRefused(t *testing.T) {
	dir := t.TempDir()
	bundlePath := filepath.Join(dir, "bundle.json")
	os.WriteFile(bundlePath, []byte(`{"test":"data"}`), 0o644)

	// Use a port that's definitely not running a registry
	_, err := PublishOCI(bundlePath, "localhost:1/test/unreachable:v1")
	if err == nil {
		t.Fatal("expected error for unreachable registry")
	}
}

// --- PullOCI: empty layers image ---
// Note: This is hard to test directly since go-containerregistry always
// creates at least one layer, but we can test pulling from a bad reference
// that resolves to an unexpected state.

func TestPullOCI_BadDigestRef(t *testing.T) {
	host := startRegistry(t)
	// Non-existent digest reference
	ref := fmt.Sprintf("%s/test/repo@sha256:0000000000000000000000000000000000000000000000000000000000000000", host)
	err := PullOCI(ref, filepath.Join(t.TempDir(), "out.json"))
	if err == nil {
		t.Fatal("expected error for non-existent digest")
	}
}

// --- SaveLocal: preserves filename from source ---

func TestSaveLocal_PreservesFilename(t *testing.T) {
	srcDir := t.TempDir()
	srcFile := filepath.Join(srcDir, "my-special-bundle.json")
	os.WriteFile(srcFile, []byte("content"), 0o644)

	dstDir := t.TempDir()
	dst, err := SaveLocal(srcFile, dstDir)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(dst) != "my-special-bundle.json" {
		t.Fatalf("expected preserved filename, got %q", filepath.Base(dst))
	}
}

// --- EnsureDefaultAttestationDir: returns correct path ---

func TestEnsureDefaultAttestationDir_ReturnsCorrectPath(t *testing.T) {
	orig, _ := os.Getwd()
	tmp := t.TempDir()
	os.Chdir(tmp)
	t.Cleanup(func() { os.Chdir(orig) })

	dir, err := EnsureDefaultAttestationDir()
	if err != nil {
		t.Fatal(err)
	}
	if dir != ".llmsa/attestations" {
		t.Fatalf("unexpected dir: %q", dir)
	}

	// Verify full path exists on disk
	fullPath := filepath.Join(tmp, ".llmsa", "attestations")
	info, err := os.Stat(fullPath)
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsDir() {
		t.Fatal("expected directory")
	}
}
