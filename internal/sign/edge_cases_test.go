package sign

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- NewPEMSigner edge cases ---

func TestNewPEMSigner_InvalidPKCS8Content(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad-pkcs8.pem")
	// Valid PEM structure but garbage DER content
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte("not-valid-pkcs8-bytes")})
	os.WriteFile(path, pemBytes, 0o600)

	_, err := NewPEMSigner(path)
	if err == nil {
		t.Fatal("expected error for invalid PKCS8 content")
	}
	if !strings.Contains(err.Error(), "parse pkcs8 key") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewPEMSigner_WrongKeyType(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ecdsa.pem")

	// Generate ECDSA key (not ed25519)
	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	pkcs8, err := x509.MarshalPKCS8PrivateKey(ecKey)
	if err != nil {
		t.Fatal(err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8})
	os.WriteFile(path, pemBytes, 0o600)

	_, err = NewPEMSigner(path)
	if err == nil {
		t.Fatal("expected error for non-ed25519 key type")
	}
	if !strings.Contains(err.Error(), "unsupported key type") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- SigstoreSigner PEM fallback error paths ---

func TestSigstoreSignerPEMFallback_InvalidKeyFile(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "bad.pem")
	os.WriteFile(keyPath, []byte("not-a-pem-key"), 0o600)

	signer := &SigstoreSigner{PEMKeyPath: keyPath}
	_, err := signer.Sign([]byte(`{"k":"v"}`))
	if err == nil {
		t.Fatal("expected error for invalid PEM key in sigstore fallback")
	}
	if !strings.Contains(err.Error(), "invalid pem key") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- publicKeyFromCertificatePEM edge cases ---

func TestPublicKeyFromCertificatePEM_InvalidPEM(t *testing.T) {
	_, _, err := publicKeyFromCertificatePEM("not-a-pem")
	if err == nil {
		t.Fatal("expected error for invalid PEM")
	}
	if !strings.Contains(err.Error(), "invalid certificate pem") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPublicKeyFromCertificatePEM_InvalidDER(t *testing.T) {
	// Valid PEM structure but garbage DER bytes
	certPEM := string(pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: []byte("not-a-valid-certificate"),
	}))
	_, _, err := publicKeyFromCertificatePEM(certPEM)
	if err == nil {
		t.Fatal("expected error for invalid certificate DER")
	}
	if !strings.Contains(err.Error(), "parse certificate") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPublicKeyFromCertificatePEM_ValidCert(t *testing.T) {
	certPEM := selfSignedCertPEM(t)
	pubPEM, digest, err := publicKeyFromCertificatePEM(certPEM)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pubPEM == "" {
		t.Fatal("expected non-empty public key PEM")
	}
	if !strings.Contains(pubPEM, "PUBLIC KEY") {
		t.Fatal("expected PUBLIC KEY header in PEM")
	}
	if digest == "" {
		t.Fatal("expected non-empty digest")
	}
}

// --- signMaterialFromCosignOutputs edge cases ---

func TestSignMaterialFromCosignOutputs_BothEmpty(t *testing.T) {
	_, err := signMaterialFromCosignOutputs("", "", "", "")
	if err == nil {
		t.Fatal("expected error for empty sig and cert")
	}
}

// --- defaultOIDCClaims edge cases ---

func TestDefaultOIDCClaims_AllDefaultsNoEnv(t *testing.T) {
	t.Setenv("GITHUB_WORKFLOW_REF", "")
	t.Setenv("GITHUB_REPOSITORY", "")
	t.Setenv("GITHUB_WORKFLOW", "")
	t.Setenv("GITHUB_REF", "")

	issuer, identity := defaultOIDCClaims("", "")
	if issuer != "https://token.actions.githubusercontent.com" {
		t.Fatalf("unexpected issuer: %s", issuer)
	}
	if !strings.Contains(identity, "local/dev") {
		t.Fatalf("expected local/dev in identity, got: %s", identity)
	}
	if !strings.Contains(identity, "manual.yml") {
		t.Fatalf("expected manual.yml in identity, got: %s", identity)
	}
}

func TestDefaultOIDCClaims_CustomValues(t *testing.T) {
	issuer, identity := defaultOIDCClaims("https://custom.issuer", "https://custom.identity")
	if issuer != "https://custom.issuer" {
		t.Fatalf("expected custom issuer, got: %s", issuer)
	}
	if identity != "https://custom.identity" {
		t.Fatalf("expected custom identity, got: %s", identity)
	}
}

// --- KMSSigner ---

func TestKMSSigner_ReturnsNotImplemented(t *testing.T) {
	signer := &KMSSigner{}
	_, err := signer.Sign([]byte("payload"))
	if err == nil {
		t.Fatal("expected error for KMS signer")
	}
	if !strings.Contains(err.Error(), "not implemented") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- CreateBundle error path ---

func TestCreateBundle_UnmarshalableStatement(t *testing.T) {
	// Channel cannot be marshaled to JSON
	_, err := CreateBundle(make(chan int), SignMaterial{})
	if err == nil {
		t.Fatal("expected error for unmarshalable statement")
	}
}

// --- DecodePayload: valid base64 but invalid JSON ---

func TestDecodePayload_ValidBase64InvalidJSON(t *testing.T) {
	bundle := Bundle{
		Envelope: Envelope{
			Payload: base64.StdEncoding.EncodeToString([]byte("not-json-content")),
		},
	}
	var out map[string]any
	err := DecodePayload(bundle, &out)
	if err == nil {
		t.Fatal("expected error for non-JSON payload")
	}
	if !strings.Contains(err.Error(), "unmarshal bundle payload") {
		t.Fatalf("unexpected error: %v", err)
	}
}
