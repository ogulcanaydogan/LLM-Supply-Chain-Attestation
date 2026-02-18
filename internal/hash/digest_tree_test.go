package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDigestTree_SingleFile(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello"), 0o644)

	digest, manifest, entries, err := DigestTree(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(entries))
	}
	if entries[0].Path != "a.txt" {
		t.Errorf("path = %q", entries[0].Path)
	}
	if !strings.HasPrefix(digest, "sha256:") {
		t.Errorf("digest should start with sha256:, got %q", digest)
	}
	if manifest == "" {
		t.Error("manifest is empty")
	}
	if !strings.Contains(manifest, "a.txt") {
		t.Error("manifest should contain filename")
	}
}

func TestDigestTree_MultipleFiles(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "c.txt"), []byte("cc"), 0o644)
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("aa"), 0o644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("bb"), 0o644)

	_, _, entries, err := DigestTree(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("entries = %d, want 3", len(entries))
	}
	// Should be sorted lexicographically
	if entries[0].Path != "a.txt" {
		t.Errorf("first entry = %q, want a.txt", entries[0].Path)
	}
	if entries[1].Path != "b.txt" {
		t.Errorf("second entry = %q, want b.txt", entries[1].Path)
	}
	if entries[2].Path != "c.txt" {
		t.Errorf("third entry = %q, want c.txt", entries[2].Path)
	}
}

func TestDigestTree_NestedDirs(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "sub", "deep"), 0o755)
	os.WriteFile(filepath.Join(dir, "root.txt"), []byte("r"), 0o644)
	os.WriteFile(filepath.Join(dir, "sub", "mid.txt"), []byte("m"), 0o644)
	os.WriteFile(filepath.Join(dir, "sub", "deep", "leaf.txt"), []byte("l"), 0o644)

	_, manifest, entries, err := DigestTree(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("entries = %d, want 3", len(entries))
	}
	// Verify forward slashes in paths
	for _, e := range entries {
		if strings.Contains(e.Path, "\\") {
			t.Errorf("path %q contains backslash", e.Path)
		}
	}
	if !strings.Contains(manifest, "sub/deep/leaf.txt") {
		t.Error("manifest should contain nested path with forward slashes")
	}
}

func TestDigestTree_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	digest, manifest, entries, err := DigestTree(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("entries = %d, want 0", len(entries))
	}
	if manifest != "" {
		t.Errorf("manifest should be empty for empty dir, got %q", manifest)
	}
	// Digest of empty string
	h := sha256.Sum256([]byte(""))
	want := "sha256:" + hex.EncodeToString(h[:])
	if digest != want {
		t.Errorf("digest = %q, want %q", digest, want)
	}
}

func TestDigestTree_DeterministicDigest(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "x.txt"), []byte("data"), 0o644)

	d1, _, _, _ := DigestTree(dir)
	d2, _, _, _ := DigestTree(dir)
	if d1 != d2 {
		t.Errorf("tree digest not deterministic: %q vs %q", d1, d2)
	}
}

func TestDigestBytes(t *testing.T) {
	data := []byte("test data")
	h := sha256.Sum256(data)
	want := "sha256:" + hex.EncodeToString(h[:])

	got := DigestBytes(data)
	if got != want {
		t.Errorf("DigestBytes = %q, want %q", got, want)
	}
}

func TestDigestBytes_Empty(t *testing.T) {
	got := DigestBytes([]byte{})
	h := sha256.Sum256([]byte{})
	want := "sha256:" + hex.EncodeToString(h[:])
	if got != want {
		t.Errorf("DigestBytes empty = %q, want %q", got, want)
	}
}

func TestFileExists_True(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "exists.txt")
	os.WriteFile(path, []byte("x"), 0o644)

	if !FileExists(path) {
		t.Error("FileExists should return true for existing file")
	}
}

func TestFileExists_False(t *testing.T) {
	if FileExists("/nonexistent/path/file.txt") {
		t.Error("FileExists should return false for nonexistent file")
	}
}
