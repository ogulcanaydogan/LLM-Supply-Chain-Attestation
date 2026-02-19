package verify

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/hash"
)

func TestVerifySubjects_AllMatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.txt")
	content := []byte("subject content")
	os.WriteFile(path, content, 0o644)

	h := sha256.Sum256(content)
	digest := hex.EncodeToString(h[:])

	statement := map[string]any{
		"subject": []any{
			map[string]any{
				"uri":    path,
				"digest": map[string]any{"sha256": digest},
			},
		},
	}

	if err := VerifySubjects(statement); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVerifySubjects_DigestMismatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.txt")
	os.WriteFile(path, []byte("original"), 0o644)

	statement := map[string]any{
		"subject": []any{
			map[string]any{
				"uri":    path,
				"digest": map[string]any{"sha256": "0000000000000000000000000000000000000000000000000000000000000000"},
			},
		},
	}

	err := VerifySubjects(statement)
	if err == nil {
		t.Fatal("expected error for digest mismatch")
	}
	if !strings.Contains(err.Error(), "digest mismatch") {
		t.Errorf("error = %q", err)
	}
}

func TestVerifySubjects_MissingFile(t *testing.T) {
	statement := map[string]any{
		"subject": []any{
			map[string]any{
				"uri":    "/nonexistent/path/file.txt",
				"digest": map[string]any{"sha256": "abc"},
			},
		},
	}

	err := VerifySubjects(statement)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "missing") {
		t.Errorf("error = %q", err)
	}
}

func TestVerifySubjects_MissingURI(t *testing.T) {
	statement := map[string]any{
		"subject": []any{
			map[string]any{
				"digest": map[string]any{"sha256": "abc"},
			},
		},
	}

	err := VerifySubjects(statement)
	if err == nil {
		t.Fatal("expected error for missing uri")
	}
	if !strings.Contains(err.Error(), "missing") {
		t.Errorf("error = %q", err)
	}
}

func TestVerifySubjects_NotArray(t *testing.T) {
	statement := map[string]any{
		"subject": "not-an-array",
	}

	err := VerifySubjects(statement)
	if err == nil {
		t.Fatal("expected error for non-array subject")
	}
	if !strings.Contains(err.Error(), "array") {
		t.Errorf("error = %q", err)
	}
}

func TestVerifySubjects_MultipleSubjects(t *testing.T) {
	dir := t.TempDir()

	subjects := make([]any, 0)
	for _, name := range []string{"a.txt", "b.txt", "c.txt"} {
		content := []byte("content of " + name)
		path := filepath.Join(dir, name)
		os.WriteFile(path, content, 0o644)
		h := sha256.Sum256(content)
		subjects = append(subjects, map[string]any{
			"uri":    path,
			"digest": map[string]any{"sha256": hex.EncodeToString(h[:])},
		})
	}

	statement := map[string]any{"subject": subjects}
	if err := VerifySubjects(statement); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVerifySubjects_DirectoryDigest(t *testing.T) {
	dir := t.TempDir()
	subDir := filepath.Join(dir, "prompts")
	os.MkdirAll(subDir, 0o755)
	os.WriteFile(filepath.Join(subDir, "a.txt"), []byte("file a"), 0o644)
	os.WriteFile(filepath.Join(subDir, "b.txt"), []byte("file b"), 0o644)

	// Compute the tree digest
	treeDigest, _, _, err := hash.DigestTree(subDir)
	if err != nil {
		t.Fatal(err)
	}
	cleanDigest := strings.TrimPrefix(treeDigest, "sha256:")

	statement := map[string]any{
		"subject": []any{
			map[string]any{
				"uri":    subDir,
				"digest": map[string]any{"sha256": cleanDigest},
			},
		},
	}

	if err := VerifySubjects(statement); err != nil {
		t.Fatalf("expected directory subject to pass: %v", err)
	}
}

func TestVerifySubjects_InvalidSubjectEntry(t *testing.T) {
	statement := map[string]any{
		"subject": []any{
			"not-a-map",
		},
	}
	err := VerifySubjects(statement)
	if err == nil {
		t.Fatal("expected error for non-map subject entry")
	}
	if !strings.Contains(err.Error(), "invalid subject entry") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVerifySubjects_EmptySubjects(t *testing.T) {
	statement := map[string]any{
		"subject": []any{},
	}
	// Empty subjects should pass (nothing to verify)
	if err := VerifySubjects(statement); err != nil {
		t.Fatalf("expected empty subjects to pass: %v", err)
	}
}
