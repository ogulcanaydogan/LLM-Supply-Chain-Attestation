package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDigestFile_KnownContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	content := []byte("hello world")
	os.WriteFile(path, content, 0o644)

	h := sha256.Sum256(content)
	want := "sha256:" + hex.EncodeToString(h[:])

	digest, size, err := DigestFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if digest != want {
		t.Errorf("digest = %q, want %q", digest, want)
	}
	if size != int64(len(content)) {
		t.Errorf("size = %d, want %d", size, len(content))
	}
}

func TestDigestFile_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.txt")
	os.WriteFile(path, []byte{}, 0o644)

	h := sha256.Sum256([]byte{})
	want := "sha256:" + hex.EncodeToString(h[:])

	digest, size, err := DigestFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if digest != want {
		t.Errorf("digest = %q, want %q", digest, want)
	}
	if size != 0 {
		t.Errorf("size = %d, want 0", size)
	}
}

func TestDigestFile_Format(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.bin")
	os.WriteFile(path, []byte{0x01, 0x02, 0x03}, 0o644)

	digest, _, err := DigestFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(digest, "sha256:") {
		t.Errorf("digest should start with 'sha256:', got %q", digest)
	}
	hexPart := strings.TrimPrefix(digest, "sha256:")
	if len(hexPart) != 64 {
		t.Errorf("hex part should be 64 chars, got %d", len(hexPart))
	}
}

func TestDigestFile_NotFound(t *testing.T) {
	_, _, err := DigestFile("/nonexistent/file.txt")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestDigestFile_SizeCorrect(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sized.bin")
	data := make([]byte, 4096)
	for i := range data {
		data[i] = byte(i % 256)
	}
	os.WriteFile(path, data, 0o644)

	_, size, err := DigestFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if size != 4096 {
		t.Errorf("size = %d, want 4096", size)
	}
}
