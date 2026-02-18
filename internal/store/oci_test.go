package store

import (
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/registry"
)

// startRegistry spins up an in-memory OCI registry and returns
// the host:port prefix suitable for use in image references.
func startRegistry(t *testing.T) string {
	t.Helper()
	handler := registry.New()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	// Strip the "http://" scheme — go-containerregistry expects host:port.
	return strings.TrimPrefix(srv.URL, "http://")
}

func TestPublishOCI_Success(t *testing.T) {
	host := startRegistry(t)

	dir := t.TempDir()
	bundlePath := filepath.Join(dir, "bundle.json")
	content := []byte(`{"envelope":{"payload":"dGVzdA==","signatures":[]}}`)
	if err := os.WriteFile(bundlePath, content, 0o644); err != nil {
		t.Fatal(err)
	}

	ref := fmt.Sprintf("%s/test/attestations:v1", host)
	pinned, err := PublishOCI(bundlePath, ref)
	if err != nil {
		t.Fatalf("PublishOCI: %v", err)
	}
	if pinned == "" {
		t.Fatal("expected non-empty pinned reference")
	}
	// Pinned reference should contain a sha256 digest.
	if !strings.Contains(pinned, "@sha256:") {
		t.Errorf("pinned = %q, want sha256 digest", pinned)
	}
}

func TestPublishOCI_InvalidRef(t *testing.T) {
	dir := t.TempDir()
	bundlePath := filepath.Join(dir, "bundle.json")
	os.WriteFile(bundlePath, []byte("{}"), 0o644)

	// Use a clearly invalid OCI reference.
	_, err := PublishOCI(bundlePath, "INVALID:::REF")
	if err == nil {
		t.Fatal("expected error for invalid OCI reference")
	}
}

func TestPublishOCI_MissingFile(t *testing.T) {
	_, err := PublishOCI("/nonexistent/bundle.json", "localhost:5000/test:v1")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestPullOCI_Success(t *testing.T) {
	host := startRegistry(t)

	// First publish a bundle.
	dir := t.TempDir()
	bundlePath := filepath.Join(dir, "bundle.json")
	content := []byte(`{"envelope":{"payloadType":"application/vnd.dsse+json","payload":"aGVsbG8=","signatures":[]}}`)
	if err := os.WriteFile(bundlePath, content, 0o644); err != nil {
		t.Fatal(err)
	}

	ref := fmt.Sprintf("%s/test/pull-test:v1", host)
	if _, err := PublishOCI(bundlePath, ref); err != nil {
		t.Fatalf("PublishOCI (setup): %v", err)
	}

	// Now pull it back.
	outPath := filepath.Join(dir, "pulled.json")
	if err := PullOCI(ref, outPath); err != nil {
		t.Fatalf("PullOCI: %v", err)
	}

	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(content) {
		t.Errorf("content mismatch:\ngot:  %q\nwant: %q", string(got), string(content))
	}
}

func TestPullOCI_InvalidRef(t *testing.T) {
	err := PullOCI("INVALID:::REF", "/tmp/out.json")
	if err == nil {
		t.Fatal("expected error for invalid OCI reference")
	}
}

func TestPullOCI_NotFound(t *testing.T) {
	host := startRegistry(t)
	ref := fmt.Sprintf("%s/nonexistent/repo:missing", host)
	err := PullOCI(ref, filepath.Join(t.TempDir(), "out.json"))
	if err == nil {
		t.Fatal("expected error for missing image")
	}
}

func TestOCIRoundTrip(t *testing.T) {
	host := startRegistry(t)

	// Create a realistic bundle payload.
	dir := t.TempDir()
	bundlePath := filepath.Join(dir, "round-trip.json")
	content := []byte(`{
  "envelope": {
    "payloadType": "application/vnd.dsse+json",
    "payload": "eyJzY2hlbWFfdmVyc2lvbiI6IjEuMC4wIn0=",
    "signatures": [{"keyid": "test-key", "sig": "dGVzdC1zaWc="}]
  },
  "metadata": {
    "bundle_version": "0.1.0",
    "created_at": "2025-01-01T00:00:00Z",
    "statement_hash": "sha256:abc123"
  }
}`)
	if err := os.WriteFile(bundlePath, content, 0o644); err != nil {
		t.Fatal(err)
	}

	ref := fmt.Sprintf("%s/org/attestations:sha256-abc123", host)
	pinned, err := PublishOCI(bundlePath, ref)
	if err != nil {
		t.Fatalf("PublishOCI: %v", err)
	}

	// Pull using the pinned digest reference.
	outPath := filepath.Join(dir, "pulled.json")
	if err := PullOCI(pinned, outPath); err != nil {
		t.Fatalf("PullOCI with pinned ref: %v", err)
	}

	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(content) {
		t.Error("round-trip content mismatch: publish then pull produced different bytes")
	}
}

func TestOCIRoundTrip_LargePayload(t *testing.T) {
	host := startRegistry(t)

	dir := t.TempDir()
	bundlePath := filepath.Join(dir, "large.json")
	// Generate a ~128KB payload to test larger bundles.
	data := make([]byte, 128*1024)
	for i := range data {
		data[i] = byte('A' + (i % 26))
	}
	if err := os.WriteFile(bundlePath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	ref := fmt.Sprintf("%s/test/large:v1", host)
	_, err := PublishOCI(bundlePath, ref)
	if err != nil {
		t.Fatalf("PublishOCI large: %v", err)
	}

	outPath := filepath.Join(dir, "pulled-large.json")
	if err := PullOCI(ref, outPath); err != nil {
		t.Fatalf("PullOCI large: %v", err)
	}

	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(data) {
		t.Errorf("size mismatch: got %d, want %d", len(got), len(data))
	}
}

func TestPublishOCI_MultipleTagsSameRepo(t *testing.T) {
	host := startRegistry(t)
	dir := t.TempDir()

	for i := 1; i <= 3; i++ {
		bundlePath := filepath.Join(dir, fmt.Sprintf("bundle%d.json", i))
		content := []byte(fmt.Sprintf(`{"id":%d}`, i))
		os.WriteFile(bundlePath, content, 0o644)

		ref := fmt.Sprintf("%s/test/multi:v%d", host, i)
		_, err := PublishOCI(bundlePath, ref)
		if err != nil {
			t.Fatalf("PublishOCI v%d: %v", i, err)
		}
	}

	// Verify we can pull each tag and get the correct content.
	for i := 1; i <= 3; i++ {
		ref := fmt.Sprintf("%s/test/multi:v%d", host, i)
		outPath := filepath.Join(dir, fmt.Sprintf("out%d.json", i))
		if err := PullOCI(ref, outPath); err != nil {
			t.Fatalf("PullOCI v%d: %v", i, err)
		}
		got, _ := os.ReadFile(outPath)
		want := fmt.Sprintf(`{"id":%d}`, i)
		if string(got) != want {
			t.Errorf("tag v%d: got %q, want %q", i, string(got), want)
		}
	}
}
