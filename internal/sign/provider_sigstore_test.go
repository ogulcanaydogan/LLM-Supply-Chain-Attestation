package sign

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"strings"
	"testing"
	"time"
)

func TestDefaultOIDCClaimsUsesWorkflowRef(t *testing.T) {
	t.Setenv("GITHUB_WORKFLOW_REF", "acme/llmsa/.github/workflows/release.yml@refs/tags/v0.1.0-rc2")
	t.Setenv("GITHUB_REPOSITORY", "")
	t.Setenv("GITHUB_WORKFLOW", "")
	t.Setenv("GITHUB_REF", "")

	issuer, identity := defaultOIDCClaims("", "")
	if issuer != "https://token.actions.githubusercontent.com" {
		t.Fatalf("unexpected issuer: %s", issuer)
	}
	wantIdentity := "https://github.com/acme/llmsa/.github/workflows/release.yml@refs/tags/v0.1.0-rc2"
	if identity != wantIdentity {
		t.Fatalf("unexpected identity: got %q want %q", identity, wantIdentity)
	}
}

func TestDefaultOIDCClaimsUsesFallbackComponents(t *testing.T) {
	t.Setenv("GITHUB_WORKFLOW_REF", "")
	t.Setenv("GITHUB_REPOSITORY", "acme/llmsa")
	t.Setenv("GITHUB_WORKFLOW", "release.yml")
	t.Setenv("GITHUB_REF", "refs/tags/v1.0.0")

	issuer, identity := defaultOIDCClaims("", "")
	if issuer != "https://token.actions.githubusercontent.com" {
		t.Fatalf("unexpected issuer: %s", issuer)
	}
	wantIdentity := "https://github.com/acme/llmsa/.github/workflows/release.yml@refs/tags/v1.0.0"
	if identity != wantIdentity {
		t.Fatalf("unexpected identity: got %q want %q", identity, wantIdentity)
	}
}

func TestSignMaterialFromCosignOutputsInvalidCertFallsBack(t *testing.T) {
	m, err := signMaterialFromCosignOutputs("base64sig", "not-a-pem-certificate", "https://issuer", "https://identity")
	if err != nil {
		t.Fatalf("expected fallback success for invalid cert text: %v", err)
	}
	if m.PublicKeyPEM != "" {
		t.Fatalf("expected empty public key on unparseable cert")
	}
	if m.CertificatePEM != "not-a-pem-certificate" {
		t.Fatalf("unexpected cert text: %q", m.CertificatePEM)
	}
	if m.SigB64 != "base64sig" {
		t.Fatalf("unexpected sig text: %q", m.SigB64)
	}
	if !strings.HasPrefix(m.KeyID, "sigstore-") {
		t.Fatalf("unexpected keyid: %q", m.KeyID)
	}
}

func TestSignMaterialFromCosignOutputsValidCertDerivesPublicKey(t *testing.T) {
	certPEM := selfSignedCertPEM(t)
	wantPub, wantDigest, err := publicKeyFromCertificatePEM(certPEM)
	if err != nil {
		t.Fatalf("test cert parse failed: %v", err)
	}

	m, err := signMaterialFromCosignOutputs("base64sig", certPEM, "https://issuer", "https://identity")
	if err != nil {
		t.Fatalf("expected success for valid cert: %v", err)
	}
	if m.PublicKeyPEM != wantPub {
		t.Fatalf("unexpected public key pem")
	}
	wantKeyID := "sigstore-" + wantDigest[:12]
	if m.KeyID != wantKeyID {
		t.Fatalf("unexpected keyid: got %q want %q", m.KeyID, wantKeyID)
	}
}

func TestSignMaterialFromCosignOutputsEmptyCertFails(t *testing.T) {
	_, err := signMaterialFromCosignOutputs("base64sig", "", "https://issuer", "https://identity")
	if err == nil {
		t.Fatalf("expected error for empty cert")
	}
}

func TestSignMaterialFromCosignOutputsEmptySigFails(t *testing.T) {
	_, err := signMaterialFromCosignOutputs("", "cert", "https://issuer", "https://identity")
	if err == nil {
		t.Fatalf("expected error for empty signature")
	}
}

func selfSignedCertPEM(t *testing.T) string {
	t.Helper()

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(1 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tpl, tpl, pub, priv)
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})))
}
