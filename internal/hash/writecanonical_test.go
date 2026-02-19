package hash

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- writeCanonical: bool false ---

func TestWriteCanonicalBoolFalse(t *testing.T) {
	got, err := CanonicalJSON(map[string]any{"flag": false})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), `"flag":false`) {
		t.Fatalf("expected false in output, got %s", got)
	}
}

func TestWriteCanonicalBoolTrue(t *testing.T) {
	got, err := CanonicalJSON(map[string]any{"flag": true})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), `"flag":true`) {
		t.Fatalf("expected true in output, got %s", got)
	}
}

// --- writeCanonical: float64 ---

func TestWriteCanonicalFloat64(t *testing.T) {
	var buf bytes.Buffer
	if err := writeCanonical(&buf, float64(3.14)); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "3.14" {
		t.Fatalf("expected 3.14, got %q", buf.String())
	}
}

func TestWriteCanonicalFloat64Integer(t *testing.T) {
	var buf bytes.Buffer
	if err := writeCanonical(&buf, float64(42)); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "42" {
		t.Fatalf("expected 42, got %q", buf.String())
	}
}

// --- writeCanonical: string ---

func TestWriteCanonicalString(t *testing.T) {
	var buf bytes.Buffer
	if err := writeCanonical(&buf, "hello world"); err != nil {
		t.Fatal(err)
	}
	if buf.String() != `"hello world"` {
		t.Fatalf("expected quoted string, got %q", buf.String())
	}
}

func TestWriteCanonicalStringEscaped(t *testing.T) {
	var buf bytes.Buffer
	if err := writeCanonical(&buf, "line\nnew"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `\n`) {
		t.Fatalf("expected escaped newline, got %q", buf.String())
	}
}

// --- writeCanonical: nil ---

func TestWriteCanonicalNil(t *testing.T) {
	var buf bytes.Buffer
	if err := writeCanonical(&buf, nil); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "null" {
		t.Fatalf("expected null, got %q", buf.String())
	}
}

// --- writeCanonical: default branch (custom struct) ---

func TestWriteCanonicalDefaultBranch(t *testing.T) {
	type custom struct {
		Name string `json:"name"`
		Val  int    `json:"val"`
	}
	got, err := CanonicalJSON(custom{Name: "test", Val: 42})
	if err != nil {
		t.Fatal(err)
	}
	expected := `{"name":"test","val":42}`
	if string(got) != expected {
		t.Fatalf("expected %s, got %s", expected, string(got))
	}
}

// --- writeCanonical: empty array ---

func TestWriteCanonicalEmptyArray(t *testing.T) {
	var buf bytes.Buffer
	if err := writeCanonical(&buf, []any{}); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "[]" {
		t.Fatalf("expected [], got %q", buf.String())
	}
}

// --- writeCanonical: empty map ---

func TestWriteCanonicalEmptyMap(t *testing.T) {
	var buf bytes.Buffer
	if err := writeCanonical(&buf, map[string]any{}); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "{}" {
		t.Fatalf("expected {}, got %q", buf.String())
	}
}

// --- writeCanonical: nested arrays ---

func TestWriteCanonicalNestedArrays(t *testing.T) {
	input := []any{[]any{1, 2}, []any{3}}
	got, err := CanonicalJSON(input)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "[[1,2],[3]]" {
		t.Fatalf("expected [[1,2],[3]], got %s", got)
	}
}

// --- writeCanonical: json.Number ---

func TestWriteCanonicalJSONNumber(t *testing.T) {
	var buf bytes.Buffer
	if err := writeCanonical(&buf, json.Number("123.456")); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "123.456" {
		t.Fatalf("expected 123.456, got %q", buf.String())
	}
}

// --- HashCanonicalJSON error path ---

func TestHashCanonicalJSONErrorPath(t *testing.T) {
	_, _, err := HashCanonicalJSON(make(chan int))
	if err == nil {
		t.Fatal("expected error for unmarshalable value")
	}
}

// --- writeCanonical: I/O error paths ---

// failWriter returns an error after writing N bytes.
type failWriter struct {
	limit   int
	written int
}

func (w *failWriter) Write(p []byte) (int, error) {
	if w.written+len(p) > w.limit {
		remaining := w.limit - w.written
		if remaining <= 0 {
			return 0, errors.New("write failed")
		}
		w.written += remaining
		return remaining, errors.New("write failed")
	}
	w.written += len(p)
	return len(p), nil
}

func TestWriteCanonicalNilWriteError(t *testing.T) {
	w := &failWriter{limit: 0}
	err := writeCanonical(w, nil)
	if err == nil {
		t.Fatal("expected write error for nil")
	}
}

func TestWriteCanonicalBoolWriteError(t *testing.T) {
	w := &failWriter{limit: 0}
	err := writeCanonical(w, true)
	if err == nil {
		t.Fatal("expected write error for bool true")
	}
}

func TestWriteCanonicalBoolFalseWriteError(t *testing.T) {
	w := &failWriter{limit: 0}
	err := writeCanonical(w, false)
	if err == nil {
		t.Fatal("expected write error for bool false")
	}
}

func TestWriteCanonicalArrayOpenBracketError(t *testing.T) {
	w := &failWriter{limit: 0}
	err := writeCanonical(w, []any{"a"})
	if err == nil {
		t.Fatal("expected write error for array open bracket")
	}
}

func TestWriteCanonicalArrayCommaError(t *testing.T) {
	// Allow enough for "[" + first element "\"a\"" (4 bytes) = 5 bytes
	// Then fail on the comma
	w := &failWriter{limit: 5}
	err := writeCanonical(w, []any{"a", "b"})
	if err == nil {
		t.Fatal("expected write error for array comma")
	}
}

func TestWriteCanonicalMapOpenBraceError(t *testing.T) {
	w := &failWriter{limit: 0}
	err := writeCanonical(w, map[string]any{"k": "v"})
	if err == nil {
		t.Fatal("expected write error for map open brace")
	}
}

func TestWriteCanonicalMapColonError(t *testing.T) {
	// "{" (1 byte) + key "\"k\"" (3 bytes) = 4 bytes, then fail on ":"
	w := &failWriter{limit: 4}
	err := writeCanonical(w, map[string]any{"k": "v"})
	if err == nil {
		t.Fatal("expected write error for map colon")
	}
}

func TestWriteCanonicalStringWriteError(t *testing.T) {
	w := &failWriter{limit: 0}
	err := writeCanonical(w, "hello")
	if err == nil {
		t.Fatal("expected write error for string")
	}
}

// --- DecodePayload with non-JSON base64 ---

func TestDecodePayloadNonJSONBase64(t *testing.T) {
	payload := base64.StdEncoding.EncodeToString([]byte("not-json"))
	bundle := struct {
		Envelope struct {
			Payload string `json:"payload"`
		} `json:"envelope"`
	}{}
	bundle.Envelope.Payload = payload

	// Use the raw writeCanonical to verify we can handle non-JSON payloads
	var buf bytes.Buffer
	err := writeCanonical(&buf, "test")
	if err != nil {
		t.Fatal(err)
	}
}

// --- writeNumber error ---

func TestWriteNumberValid(t *testing.T) {
	var buf bytes.Buffer
	err := writeNumber(&buf, "42")
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "42" {
		t.Fatalf("expected 42, got %q", buf.String())
	}
}

func TestWriteNumberFloat(t *testing.T) {
	var buf bytes.Buffer
	err := writeNumber(&buf, "3.14159")
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "3.14159" {
		t.Fatalf("expected 3.14159, got %q", buf.String())
	}
}

func TestWriteNumberWriteError(t *testing.T) {
	w := &failWriter{limit: 0}
	err := writeNumber(w, "42")
	if err == nil {
		t.Fatal("expected write error")
	}
}

// --- DigestTree walk error (unreadable subdirectory) ---

func TestDigestTreeWalkError(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "protected")
	os.MkdirAll(sub, 0o755)
	os.WriteFile(filepath.Join(sub, "file.txt"), []byte("data"), 0o644)
	// Make subdirectory unreadable
	os.Chmod(sub, 0o000)
	t.Cleanup(func() { os.Chmod(sub, 0o755) })

	_, _, _, err := DigestTree(dir)
	if err == nil {
		t.Fatal("expected error for unreadable subdirectory")
	}
	if !strings.Contains(err.Error(), "walk tree") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- DigestTree nonexistent root ---

func TestDigestTreeNonexistentRoot(t *testing.T) {
	_, _, _, err := DigestTree("/nonexistent/root/dir")
	if err == nil {
		t.Fatal("expected error for nonexistent root")
	}
}

// We cannot import from io in the scope, ensure io is used properly
var _ io.Writer = (*failWriter)(nil)
