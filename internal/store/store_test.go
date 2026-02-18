package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLocal(t *testing.T) {
	srcDir := t.TempDir()
	srcFile := filepath.Join(srcDir, "bundle.json")
	content := []byte(`{"envelope":{"payload":"test"}}`)
	if err := os.WriteFile(srcFile, content, 0o644); err != nil {
		t.Fatal(err)
	}

	dstDir := filepath.Join(t.TempDir(), "output")
	dst, err := SaveLocal(srcFile, dstDir)
	if err != nil {
		t.Fatalf("SaveLocal: %v", err)
	}

	if filepath.Base(dst) != "bundle.json" {
		t.Errorf("dest filename = %q, want bundle.json", filepath.Base(dst))
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(content) {
		t.Errorf("content mismatch: got %q", string(got))
	}
}

func TestSaveLocal_CreatesDirectory(t *testing.T) {
	srcDir := t.TempDir()
	srcFile := filepath.Join(srcDir, "stmt.json")
	if err := os.WriteFile(srcFile, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	dstDir := filepath.Join(t.TempDir(), "a", "b", "c")
	_, err := SaveLocal(srcFile, dstDir)
	if err != nil {
		t.Fatalf("SaveLocal with nested dir: %v", err)
	}

	info, err := os.Stat(dstDir)
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsDir() {
		t.Error("expected directory to be created")
	}
}

func TestSaveLocal_SourceNotFound(t *testing.T) {
	_, err := SaveLocal("/nonexistent/file.json", t.TempDir())
	if err == nil {
		t.Fatal("expected error for missing source")
	}
}

func TestEnsureDefaultAttestationDir(t *testing.T) {
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)

	dir, err := EnsureDefaultAttestationDir()
	if err != nil {
		t.Fatalf("EnsureDefaultAttestationDir: %v", err)
	}
	if dir != ".llmsa/attestations" {
		t.Errorf("dir = %q", dir)
	}

	info, err := os.Stat(filepath.Join(tmp, ".llmsa", "attestations"))
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsDir() {
		t.Error("expected directory")
	}
}

func TestEnsureDefaultAttestationDir_Idempotent(t *testing.T) {
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)

	_, err = EnsureDefaultAttestationDir()
	if err != nil {
		t.Fatal(err)
	}
	_, err = EnsureDefaultAttestationDir()
	if err != nil {
		t.Fatal("second call should succeed (idempotent)")
	}
}

func TestSaveLocal_PreservesContent(t *testing.T) {
	srcDir := t.TempDir()
	srcFile := filepath.Join(srcDir, "large.json")

	data := make([]byte, 64*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	if err := os.WriteFile(srcFile, data, 0o644); err != nil {
		t.Fatal(err)
	}

	dstDir := t.TempDir()
	dst, err := SaveLocal(srcFile, dstDir)
	if err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(data) {
		t.Errorf("size mismatch: got %d, want %d", len(got), len(data))
	}
	for i := range data {
		if got[i] != data[i] {
			t.Errorf("byte mismatch at offset %d", i)
			break
		}
	}
}
