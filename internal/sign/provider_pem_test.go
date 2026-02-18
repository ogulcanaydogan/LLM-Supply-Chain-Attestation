package sign

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGeneratePEMPrivateKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.pem")

	if err := GeneratePEMPrivateKey(path); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("permissions = %o, want 600", info.Mode().Perm())
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	block, _ := pem.Decode(raw)
	if block == nil {
		t.Fatal("not valid PEM")
	}
	if block.Type != "PRIVATE KEY" {
		t.Errorf("PEM type = %q, want PRIVATE KEY", block.Type)
	}
}

func TestNewPEMSigner(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "key.pem")
	GeneratePEMPrivateKey(path)

	signer, err := NewPEMSigner(path)
	if err != nil {
		t.Fatal(err)
	}
	if signer.PublicKey == nil {
		t.Error("public key is nil")
	}
	if signer.PrivateKey == nil {
		t.Error("private key is nil")
	}
}

func TestNewPEMSigner_InvalidPEM(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.pem")
	os.WriteFile(path, []byte("not a pem file"), 0o600)

	_, err := NewPEMSigner(path)
	if err == nil {
		t.Fatal("expected error for invalid PEM")
	}
	if !strings.Contains(err.Error(), "invalid pem") {
		t.Errorf("error = %q", err)
	}
}

func TestNewPEMSigner_FileNotFound(t *testing.T) {
	_, err := NewPEMSigner("/nonexistent/key.pem")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestPEMSigner_Sign(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "key.pem")
	GeneratePEMPrivateKey(path)
	signer, _ := NewPEMSigner(path)

	payload := []byte(`{"test":"data"}`)
	mat, err := signer.Sign(payload)
	if err != nil {
		t.Fatal(err)
	}
	if mat.SigB64 == "" {
		t.Error("signature is empty")
	}
	if mat.Provider != "pem" {
		t.Errorf("provider = %q, want pem", mat.Provider)
	}
	if mat.KeyID == "" {
		t.Error("keyid is empty")
	}
	if len(mat.KeyID) != 16 {
		t.Errorf("keyid length = %d, want 16 (8 bytes hex)", len(mat.KeyID))
	}
	if mat.PublicKeyPEM == "" {
		t.Error("public_key_pem is empty")
	}
	if !strings.Contains(mat.PublicKeyPEM, "PUBLIC KEY") {
		t.Error("public_key_pem should contain PUBLIC KEY header")
	}
}

func TestPEMSigner_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "key.pem")
	GeneratePEMPrivateKey(path)
	signer, _ := NewPEMSigner(path)

	payload := []byte(`{"statement":"test"}`)
	mat, err := signer.Sign(payload)
	if err != nil {
		t.Fatal(err)
	}

	// Decode the public key from the PEM output
	block, _ := pem.Decode([]byte(mat.PublicKeyPEM))
	if block == nil {
		t.Fatal("cannot decode public key PEM")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	pub, ok := pubInterface.(ed25519.PublicKey)
	if !ok {
		t.Fatal("not an ed25519 public key")
	}

	// Decode and verify the signature
	sigBytes, err := base64.StdEncoding.DecodeString(mat.SigB64)
	if err != nil {
		t.Fatal(err)
	}
	if !ed25519.Verify(pub, payload, sigBytes) {
		t.Error("signature verification failed")
	}
}

func TestPEMSigner_DifferentPayloads(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "key.pem")
	GeneratePEMPrivateKey(path)
	signer, _ := NewPEMSigner(path)

	mat1, _ := signer.Sign([]byte("payload-1"))
	mat2, _ := signer.Sign([]byte("payload-2"))

	if mat1.SigB64 == mat2.SigB64 {
		t.Error("different payloads should produce different signatures")
	}
	if mat1.KeyID != mat2.KeyID {
		t.Error("same key should produce same keyid")
	}
}
