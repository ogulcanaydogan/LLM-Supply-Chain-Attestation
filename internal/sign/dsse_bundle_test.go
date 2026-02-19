package sign

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCreateBundle(t *testing.T) {
	statement := map[string]any{
		"schema_version":   "1.0.0",
		"statement_id":     "stmt-test",
		"attestation_type": "prompt_attestation",
	}
	material := SignMaterial{
		KeyID:        "abc123",
		SigB64:       base64.StdEncoding.EncodeToString([]byte("fake-sig")),
		Provider:     "pem",
		PublicKeyPEM: "-----BEGIN PUBLIC KEY-----\ntest\n-----END PUBLIC KEY-----",
	}

	bundle, err := CreateBundle(statement, material)
	if err != nil {
		t.Fatal(err)
	}
	if bundle.Envelope.PayloadType != "application/vnd.llmsa.statement.v1+json" {
		t.Errorf("payloadType = %q", bundle.Envelope.PayloadType)
	}
	if bundle.Envelope.Payload == "" {
		t.Error("payload is empty")
	}
	if len(bundle.Envelope.Signatures) != 1 {
		t.Fatalf("signatures count = %d", len(bundle.Envelope.Signatures))
	}
	sig := bundle.Envelope.Signatures[0]
	if sig.KeyID != "abc123" {
		t.Errorf("keyid = %q", sig.KeyID)
	}
	if sig.Provider != "pem" {
		t.Errorf("provider = %q", sig.Provider)
	}
}

func TestCreateBundle_MetadataFields(t *testing.T) {
	statement := map[string]any{"id": "test"}
	material := SignMaterial{KeyID: "k", SigB64: "s", Provider: "pem", PublicKeyPEM: "p"}

	bundle, err := CreateBundle(statement, material)
	if err != nil {
		t.Fatal(err)
	}
	if bundle.Metadata.BundleVersion != "1" {
		t.Errorf("bundle_version = %q, want 1", bundle.Metadata.BundleVersion)
	}
	if !strings.HasPrefix(bundle.Metadata.StatementHash, "sha256:") {
		t.Errorf("statement_hash should start with sha256:, got %q", bundle.Metadata.StatementHash)
	}
	_, err = time.Parse(time.RFC3339, bundle.Metadata.CreatedAt)
	if err != nil {
		t.Errorf("created_at is not RFC3339: %q", bundle.Metadata.CreatedAt)
	}
}

func TestDecodePayload(t *testing.T) {
	statement := map[string]any{
		"statement_id":     "stmt-1",
		"attestation_type": "prompt_attestation",
	}
	material := SignMaterial{KeyID: "k", SigB64: "s", Provider: "pem", PublicKeyPEM: "p"}

	bundle, err := CreateBundle(statement, material)
	if err != nil {
		t.Fatal(err)
	}

	var decoded map[string]any
	if err := DecodePayload(bundle, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["statement_id"] != "stmt-1" {
		t.Errorf("decoded statement_id = %v", decoded["statement_id"])
	}
	if decoded["attestation_type"] != "prompt_attestation" {
		t.Errorf("decoded attestation_type = %v", decoded["attestation_type"])
	}
}

func TestDecodePayload_InvalidBase64(t *testing.T) {
	bundle := Bundle{
		Envelope: Envelope{
			Payload: "not-valid-base64!!!",
		},
	}
	var out map[string]any
	err := DecodePayload(bundle, &out)
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

func TestWriteReadBundle(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.bundle.json")

	statement := map[string]any{"id": "round-trip"}
	material := SignMaterial{KeyID: "k1", SigB64: "sig", Provider: "pem", PublicKeyPEM: "pub"}
	original, err := CreateBundle(statement, material)
	if err != nil {
		t.Fatal(err)
	}

	if err := WriteBundle(path, original); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Error("written file is empty")
	}

	loaded, err := ReadBundle(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Envelope.PayloadType != original.Envelope.PayloadType {
		t.Errorf("payloadType mismatch: %q vs %q", loaded.Envelope.PayloadType, original.Envelope.PayloadType)
	}
	if loaded.Envelope.Payload != original.Envelope.Payload {
		t.Error("payload mismatch after round-trip")
	}
	if len(loaded.Envelope.Signatures) != 1 {
		t.Errorf("signatures count = %d", len(loaded.Envelope.Signatures))
	}
	if loaded.Metadata.BundleVersion != original.Metadata.BundleVersion {
		t.Error("bundle_version mismatch")
	}
	if loaded.Metadata.StatementHash != original.Metadata.StatementHash {
		t.Error("statement_hash mismatch")
	}
}

func TestReadBundle_FileNotFound(t *testing.T) {
	_, err := ReadBundle("/nonexistent/bundle.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestReadBundle_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	os.WriteFile(path, []byte("not json"), 0o644)

	_, err := ReadBundle(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestCreateBundle_SigstoreProvider(t *testing.T) {
	statement := map[string]any{"id": "sigstore-test"}
	material := SignMaterial{
		KeyID:          "sigstore-abc123",
		SigB64:         "c2lnbmF0dXJl",
		Provider:       "sigstore",
		PublicKeyPEM:   "-----BEGIN PUBLIC KEY-----\ntest\n-----END PUBLIC KEY-----",
		CertificatePEM: "-----BEGIN CERTIFICATE-----\ncert\n-----END CERTIFICATE-----",
		OIDCIssuer:     "https://token.actions.githubusercontent.com",
		OIDCIdentity:   "https://github.com/acme/llmsa/.github/workflows/ci.yml@refs/heads/main",
	}

	bundle, err := CreateBundle(statement, material)
	if err != nil {
		t.Fatal(err)
	}
	sig := bundle.Envelope.Signatures[0]
	if sig.Provider != "sigstore" {
		t.Errorf("provider = %q, want sigstore", sig.Provider)
	}
	if sig.CertificatePEM == "" {
		t.Error("expected certificate PEM to be set")
	}
	if sig.OIDCIssuer != material.OIDCIssuer {
		t.Errorf("oidc_issuer = %q, want %q", sig.OIDCIssuer, material.OIDCIssuer)
	}
	if sig.OIDCIdentity != material.OIDCIdentity {
		t.Errorf("oidc_identity = %q, want %q", sig.OIDCIdentity, material.OIDCIdentity)
	}
}

func TestWriteBundle_InvalidPath(t *testing.T) {
	bundle := Bundle{
		Envelope: Envelope{PayloadType: "test"},
		Metadata: Metadata{BundleVersion: "1"},
	}
	err := WriteBundle("/nonexistent/dir/bundle.json", bundle)
	if err == nil {
		t.Fatal("expected error for invalid directory path")
	}
}

func TestCreateBundle_PayloadIsBase64Decodable(t *testing.T) {
	statement := map[string]any{"key": "value", "nested": map[string]any{"a": 1}}
	material := SignMaterial{KeyID: "k", SigB64: "s", Provider: "pem", PublicKeyPEM: "p"}

	bundle, err := CreateBundle(statement, material)
	if err != nil {
		t.Fatal(err)
	}

	decoded, err := base64.StdEncoding.DecodeString(bundle.Envelope.Payload)
	if err != nil {
		t.Fatalf("payload is not valid base64: %v", err)
	}
	if len(decoded) == 0 {
		t.Error("decoded payload is empty")
	}
	if !strings.Contains(string(decoded), "key") {
		t.Error("decoded payload does not contain expected content")
	}
}
