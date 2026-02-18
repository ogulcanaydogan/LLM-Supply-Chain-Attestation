package verify

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/hash"
	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/sign"
)

func TestVerifyIdentityPolicyAcceptsGitHubWorkflowIdentity(t *testing.T) {
	sig := sign.Signature{
		OIDCIssuer:   "https://token.actions.githubusercontent.com",
		OIDCIdentity: "https://github.com/acme/llmsa/.github/workflows/ci-attest-verify.yml@refs/heads/main",
	}
	policy := SignerPolicy{
		OIDCIssuer:    "https://token.actions.githubusercontent.com",
		IdentityRegex: `^https://github\.com/.+/.+/.github/workflows/.+@refs/.+$`,
	}
	if err := verifyIdentityPolicy(sig, policy); err != nil {
		t.Fatalf("expected identity to pass policy: %v", err)
	}
}

func TestVerifyIdentityPolicyRejectsMismatchedIdentity(t *testing.T) {
	sig := sign.Signature{
		OIDCIssuer:   "https://token.actions.githubusercontent.com",
		OIDCIdentity: "repo:acme/llmsa:workflow:ci",
	}
	policy := SignerPolicy{
		OIDCIssuer:    "https://token.actions.githubusercontent.com",
		IdentityRegex: `^https://github\.com/.+/.+/.github/workflows/.+@refs/.+$`,
	}
	err := verifyIdentityPolicy(sig, policy)
	if err == nil {
		t.Fatalf("expected identity mismatch error")
	}
	if !strings.Contains(err.Error(), "oidc identity mismatch") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVerifySignatureRejectsMissingSignatures(t *testing.T) {
	bundle := sign.Bundle{
		Envelope: sign.Envelope{
			PayloadType: "application/json",
			Payload:     "e30=",
			Signatures:  []sign.Signature{},
		},
	}
	if err := VerifySignature(bundle, SignerPolicy{}); err == nil {
		t.Fatal("expected signature verification error")
	}
}

func TestVerifySignatureRejectsStatementHashMismatch(t *testing.T) {
	statement := map[string]any{
		"schema_version":   "1.0.0",
		"statement_id":     "sig-1",
		"attestation_type": "prompt_attestation",
		"predicate_type":   "https://llmsa.dev/attestation/prompt/v1",
		"generated_at":     "2026-02-18T00:00:00Z",
		"generator":        map[string]any{"name": "llmsa", "version": "1.0.0", "git_sha": "abc"},
		"subject": []any{
			map[string]any{"name": "x", "uri": "/tmp/none", "digest": map[string]any{"sha256": strings.Repeat("a", 64)}, "size_bytes": 0},
		},
		"predicate": map[string]any{
			"prompt_bundle_digest": "sha256:x",
			"system_prompt_digest": "sha256:y",
			"template_digests":     []any{"sha256:z"},
			"tool_schema_digests":  []any{"sha256:w"},
			"safety_policy_digest": "sha256:v",
		},
		"privacy": map[string]any{"mode": "hash_only"},
	}
	keyPath := filepath.Join(t.TempDir(), "dev.pem")
	if err := sign.GeneratePEMPrivateKey(keyPath); err != nil {
		t.Fatal(err)
	}
	signer, err := sign.NewPEMSigner(keyPath)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := hash.CanonicalJSON(statement)
	if err != nil {
		t.Fatal(err)
	}
	material, err := signer.Sign(canonical)
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := sign.CreateBundle(statement, material)
	if err != nil {
		t.Fatal(err)
	}
	bundle.Metadata.StatementHash = "sha256:" + strings.Repeat("0", 64)
	if err := VerifySignature(bundle, SignerPolicy{}); err == nil {
		t.Fatal("expected statement hash mismatch error")
	}
}

func TestVerifySignatureSigstoreWithoutCosign(t *testing.T) {
	t.Setenv("PATH", "")
	bundle := sign.Bundle{
		Envelope: sign.Envelope{
			PayloadType: "application/json",
			Payload:     "e30=",
			Signatures: []sign.Signature{
				{
					Provider:       "sigstore",
					Sig:            "ZHVtbXk=",
					PublicKeyPEM:   "",
					CertificatePEM: "-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----",
				},
			},
		},
		Metadata: sign.Metadata{
			StatementHash: hash.DigestBytes([]byte("{}")),
		},
	}
	err := VerifySignature(bundle, SignerPolicy{})
	if err == nil {
		t.Fatal("expected cosign availability error")
	}
	if !strings.Contains(err.Error(), "cosign binary is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParsePublicKeyRejectsUnsupportedType(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	pkix, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	pubPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pkix}))
	_, err = parsePublicKey(pubPEM)
	if err == nil {
		t.Fatal("expected unsupported key type error")
	}
	if !strings.Contains(err.Error(), "unsupported public key type") {
		t.Fatalf("unexpected error: %v", err)
	}
}
