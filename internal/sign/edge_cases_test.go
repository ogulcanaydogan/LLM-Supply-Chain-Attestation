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

// --- SigstoreSigner PEM fallback: nonexistent key file ---

func TestSigstoreSignerPEMFallback_NonexistentKeyFile(t *testing.T) {
	signer := &SigstoreSigner{PEMKeyPath: "/nonexistent/key.pem"}
	_, err := signer.Sign([]byte(`{"k":"v"}`))
	if err == nil {
		t.Fatal("expected error for nonexistent PEM key path")
	}
	if !strings.Contains(err.Error(), "read pem key") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- SigstoreSigner PEM fallback: happy path ---

func TestSigstoreSignerPEMFallback_HappyPath(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "key.pem")
	if err := GeneratePEMPrivateKey(keyPath); err != nil {
		t.Fatal(err)
	}

	signer := &SigstoreSigner{
		PEMKeyPath: keyPath,
		Issuer:     "https://custom.issuer",
		Identity:   "https://custom.identity",
	}
	material, err := signer.Sign([]byte(`{"attestation":"test"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if material.Provider != "sigstore" {
		t.Fatalf("expected provider=sigstore, got %q", material.Provider)
	}
	if material.OIDCIssuer != "https://custom.issuer" {
		t.Fatalf("expected custom issuer, got %q", material.OIDCIssuer)
	}
	if material.OIDCIdentity != "https://custom.identity" {
		t.Fatalf("expected custom identity, got %q", material.OIDCIdentity)
	}
	if material.SigB64 == "" {
		t.Fatal("expected non-empty signature")
	}
	if material.PublicKeyPEM == "" {
		t.Fatal("expected non-empty public key PEM")
	}
}

// --- SigstoreSigner keyless: cosign not found ---

func TestSigstoreSignerKeyless_CosignNotFound(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	signer := &SigstoreSigner{} // No PEMKeyPath → keyless path
	_, err := signer.Sign([]byte(`{"k":"v"}`))
	if err == nil {
		t.Fatal("expected error when cosign is not in PATH")
	}
	if !strings.Contains(err.Error(), "cosign binary not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- defaultOIDCClaims: GITHUB_WORKFLOW_REF set ---

func TestDefaultOIDCClaims_GitHubWorkflowRef(t *testing.T) {
	t.Setenv("GITHUB_WORKFLOW_REF", "org/repo/.github/workflows/ci.yml@refs/heads/main")
	issuer, identity := defaultOIDCClaims("", "")
	if issuer != "https://token.actions.githubusercontent.com" {
		t.Fatalf("unexpected issuer: %s", issuer)
	}
	if identity != "https://github.com/org/repo/.github/workflows/ci.yml@refs/heads/main" {
		t.Fatalf("unexpected identity: %s", identity)
	}
}

// --- defaultOIDCClaims: individual env vars ---

func TestDefaultOIDCClaims_IndividualEnvVars(t *testing.T) {
	t.Setenv("GITHUB_WORKFLOW_REF", "")
	t.Setenv("GITHUB_REPOSITORY", "myorg/myrepo")
	t.Setenv("GITHUB_WORKFLOW", "release.yml")
	t.Setenv("GITHUB_REF", "refs/tags/v1.0.0")

	_, identity := defaultOIDCClaims("", "")
	if !strings.Contains(identity, "myorg/myrepo") {
		t.Fatalf("expected repo in identity, got: %s", identity)
	}
	if !strings.Contains(identity, "release.yml") {
		t.Fatalf("expected workflow in identity, got: %s", identity)
	}
	if !strings.Contains(identity, "refs/tags/v1.0.0") {
		t.Fatalf("expected ref in identity, got: %s", identity)
	}
}

// --- signMaterialFromCosignOutputs: valid cert ---

func TestSignMaterialFromCosignOutputs_ValidCert(t *testing.T) {
	certPEM := selfSignedCertPEM(t)
	material, err := signMaterialFromCosignOutputs("c2lnbmF0dXJl", certPEM, "https://issuer", "https://identity")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if material.Provider != "sigstore" {
		t.Fatalf("expected provider=sigstore, got %q", material.Provider)
	}
	if !strings.HasPrefix(material.KeyID, "sigstore-") {
		t.Fatalf("expected sigstore- prefix in key ID, got %q", material.KeyID)
	}
	if material.PublicKeyPEM == "" {
		t.Fatal("expected non-empty public key PEM extracted from cert")
	}
	if material.CertificatePEM != certPEM {
		t.Fatal("certificate PEM mismatch")
	}
}

// --- signMaterialFromCosignOutputs: unparseable cert ---

func TestSignMaterialFromCosignOutputs_UnparseableCert(t *testing.T) {
	badCert := "-----BEGIN CERTIFICATE-----\nYWJj\n-----END CERTIFICATE-----"
	material, err := signMaterialFromCosignOutputs("c2lnbmF0dXJl", badCert, "https://issuer", "https://identity")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if material.PublicKeyPEM != "" {
		t.Fatal("expected empty public key PEM for unparseable cert")
	}
	if !strings.HasPrefix(material.KeyID, "sigstore-") {
		t.Fatalf("expected sigstore- key ID from raw cert digest, got %q", material.KeyID)
	}
}

// --- signMaterialFromCosignOutputs: empty sig only ---

func TestSignMaterialFromCosignOutputs_EmptySigOnly(t *testing.T) {
	_, err := signMaterialFromCosignOutputs("", "some-cert", "", "")
	if err == nil {
		t.Fatal("expected error for empty signature")
	}
	if !strings.Contains(err.Error(), "empty signature") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- signMaterialFromCosignOutputs: empty cert only ---

func TestSignMaterialFromCosignOutputs_EmptyCertOnly(t *testing.T) {
	_, err := signMaterialFromCosignOutputs("sig-data", "", "", "")
	if err == nil {
		t.Fatal("expected error for empty certificate")
	}
	if !strings.Contains(err.Error(), "empty certificate") {
		t.Fatalf("unexpected error: %v", err)
	}
}
