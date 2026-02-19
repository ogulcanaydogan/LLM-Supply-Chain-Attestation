package attest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"filippo.io/age"
	"github.com/ogulcanaydogan/llm-supply-chain-attestation/pkg/types"
)

func TestApplyPrivacyConfigDefaultsToHashOnly(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "cfg.yaml")
	if err := os.WriteFile(cfgPath, []byte("system_prompt: x"), 0o644); err != nil {
		t.Fatal(err)
	}
	stmt := types.Statement{Privacy: types.Privacy{Mode: "hash_only"}}
	if err := applyPrivacyConfig(&stmt, cfgPath); err != nil {
		t.Fatal(err)
	}
	if stmt.Privacy.Mode != "hash_only" {
		t.Fatalf("expected hash_only, got %q", stmt.Privacy.Mode)
	}
}

func TestApplyPrivacyConfigEncryptedPayload(t *testing.T) {
	tmp := t.TempDir()
	secretPath := filepath.Join(tmp, "secret.txt")
	secret := "top-secret-prompt-content"
	if err := os.WriteFile(secretPath, []byte(secret), 0o600); err != nil {
		t.Fatal(err)
	}
	id, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	cfgPath := filepath.Join(tmp, "cfg.yaml")
	cfg := "privacy_mode: encrypted_payload\n" +
		"encrypted_payload_path: secret.txt\n" +
		"age_recipient: " + id.Recipient().String() + "\n"
	if err := os.WriteFile(cfgPath, []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}

	stmt := types.Statement{}
	if err := applyPrivacyConfig(&stmt, cfgPath); err != nil {
		t.Fatal(err)
	}
	if stmt.Privacy.Mode != "encrypted_payload" {
		t.Fatalf("expected encrypted_payload, got %q", stmt.Privacy.Mode)
	}
	if stmt.Privacy.EncryptedBlobDigest == "" {
		t.Fatalf("expected encrypted blob digest")
	}
	if stmt.Privacy.EncryptionRecipientFingerprint == "" {
		t.Fatalf("expected recipient fingerprint")
	}

	raw, err := json.Marshal(stmt)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), secret) {
		t.Fatalf("statement must not contain plaintext secret")
	}
}

func TestApplyPrivacyConfigRejectsInvalidEncryptedConfig(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "cfg.yaml")
	if err := os.WriteFile(cfgPath, []byte("privacy_mode: encrypted_payload\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stmt := types.Statement{}
	if err := applyPrivacyConfig(&stmt, cfgPath); err == nil {
		t.Fatalf("expected error for incomplete encrypted config")
	}
}

func TestApplyPrivacyConfigPlaintextExplicit(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "cfg.yaml")
	if err := os.WriteFile(cfgPath, []byte("privacy_mode: plaintext_explicit\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stmt := types.Statement{}
	if err := applyPrivacyConfig(&stmt, cfgPath); err != nil {
		t.Fatal(err)
	}
	if stmt.Privacy.Mode != "plaintext_explicit" {
		t.Fatalf("expected plaintext_explicit, got %q", stmt.Privacy.Mode)
	}
}

func TestApplyPrivacyConfigUnsupportedMode(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "cfg.yaml")
	if err := os.WriteFile(cfgPath, []byte("privacy_mode: unknown_mode\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stmt := types.Statement{}
	err := applyPrivacyConfig(&stmt, cfgPath)
	if err == nil {
		t.Fatal("expected error for unsupported privacy mode")
	}
	if !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplyPrivacyConfigEncryptedMissingRecipient(t *testing.T) {
	tmp := t.TempDir()
	secretPath := filepath.Join(tmp, "secret.txt")
	os.WriteFile(secretPath, []byte("data"), 0o600)
	cfgPath := filepath.Join(tmp, "cfg.yaml")
	cfg := "privacy_mode: encrypted_payload\nencrypted_payload_path: secret.txt\n"
	os.WriteFile(cfgPath, []byte(cfg), 0o644)
	stmt := types.Statement{}
	err := applyPrivacyConfig(&stmt, cfgPath)
	if err == nil {
		t.Fatal("expected error for missing age_recipient")
	}
	if !strings.Contains(err.Error(), "age_recipient") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplyPrivacyConfigEncryptedInvalidRecipient(t *testing.T) {
	tmp := t.TempDir()
	secretPath := filepath.Join(tmp, "secret.txt")
	os.WriteFile(secretPath, []byte("data"), 0o600)
	cfgPath := filepath.Join(tmp, "cfg.yaml")
	cfg := "privacy_mode: encrypted_payload\nencrypted_payload_path: secret.txt\nage_recipient: not-a-valid-recipient\n"
	os.WriteFile(cfgPath, []byte(cfg), 0o644)
	stmt := types.Statement{}
	err := applyPrivacyConfig(&stmt, cfgPath)
	if err == nil {
		t.Fatal("expected error for invalid age_recipient")
	}
	if !strings.Contains(err.Error(), "parse age_recipient") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplyPrivacyConfigEncryptedMissingPayloadFile(t *testing.T) {
	tmp := t.TempDir()
	id, _ := age.GenerateX25519Identity()
	cfgPath := filepath.Join(tmp, "cfg.yaml")
	cfg := "privacy_mode: encrypted_payload\nencrypted_payload_path: nonexistent.txt\nage_recipient: " + id.Recipient().String() + "\n"
	os.WriteFile(cfgPath, []byte(cfg), 0o644)
	stmt := types.Statement{}
	err := applyPrivacyConfig(&stmt, cfgPath)
	if err == nil {
		t.Fatal("expected error for missing payload file")
	}
	if !strings.Contains(err.Error(), "read encrypted payload") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplyPrivacyConfigEncryptedWithFingerprint(t *testing.T) {
	tmp := t.TempDir()
	secretPath := filepath.Join(tmp, "secret.txt")
	os.WriteFile(secretPath, []byte("data"), 0o600)
	id, _ := age.GenerateX25519Identity()
	cfgPath := filepath.Join(tmp, "cfg.yaml")
	cfg := "privacy_mode: encrypted_payload\nencrypted_payload_path: secret.txt\n" +
		"age_recipient: " + id.Recipient().String() + "\n" +
		"encryption_recipient_fingerprint: custom-fingerprint\n"
	os.WriteFile(cfgPath, []byte(cfg), 0o644)
	stmt := types.Statement{}
	if err := applyPrivacyConfig(&stmt, cfgPath); err != nil {
		t.Fatal(err)
	}
	if stmt.Privacy.EncryptionRecipientFingerprint != "custom-fingerprint" {
		t.Fatalf("expected custom fingerprint, got %q", stmt.Privacy.EncryptionRecipientFingerprint)
	}
}
