package verify

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/hash"
	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/sign"
)

type SignerPolicy struct {
	OIDCIssuer    string
	IdentityRegex string
}

func VerifySignature(bundle sign.Bundle, policy SignerPolicy) error {
	if len(bundle.Envelope.Signatures) == 0 {
		return fmt.Errorf("no signatures in bundle")
	}
	rawPayload, err := base64.StdEncoding.DecodeString(bundle.Envelope.Payload)
	if err != nil {
		return fmt.Errorf("decode payload: %w", err)
	}
	if hash.DigestBytes(rawPayload) != bundle.Metadata.StatementHash {
		return fmt.Errorf("statement hash mismatch")
	}

	sig := bundle.Envelope.Signatures[0]
	if sig.Provider == "sigstore" && strings.TrimSpace(sig.CertificatePEM) != "" {
		if err := verifyWithCosign(rawPayload, sig, policy); err != nil {
			return err
		}
		return nil
	}

	pub, err := parsePublicKey(sig.PublicKeyPEM)
	if err != nil {
		return err
	}
	rawSig, err := base64.StdEncoding.DecodeString(sig.Sig)
	if err != nil {
		return fmt.Errorf("decode signature: %w", err)
	}
	if !ed25519.Verify(pub, rawPayload, rawSig) {
		return fmt.Errorf("signature verification failed")
	}

	if sig.Provider == "sigstore" {
		if err := verifyIdentityPolicy(sig, policy); err != nil {
			return err
		}
	}
	return nil
}

func verifyWithCosign(payload []byte, sig sign.Signature, policy SignerPolicy) error {
	if _, err := exec.LookPath("cosign"); err != nil {
		return fmt.Errorf("cosign binary is required to verify sigstore keyless bundles: %w", err)
	}

	tmp, err := os.MkdirTemp("", "llmsa-sigstore-verify-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	payloadPath := filepath.Join(tmp, "payload.json")
	sigPath := filepath.Join(tmp, "payload.sig")
	certPath := filepath.Join(tmp, "payload.pem")
	if err := os.WriteFile(payloadPath, payload, 0o600); err != nil {
		return err
	}
	if err := os.WriteFile(sigPath, []byte(strings.TrimSpace(sig.Sig)+"\n"), 0o600); err != nil {
		return err
	}
	if err := os.WriteFile(certPath, []byte(sig.CertificatePEM), 0o600); err != nil {
		return err
	}

	args := []string{"verify-blob", "--signature", sigPath, "--certificate", certPath}
	if policy.OIDCIssuer != "" {
		args = append(args, "--certificate-oidc-issuer", policy.OIDCIssuer)
	}
	if policy.IdentityRegex != "" {
		args = append(args, "--certificate-identity-regexp", policy.IdentityRegex)
	}
	args = append(args, payloadPath)

	cmd := exec.Command("cosign", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("sigstore verification failed (including Rekor/tlog checks): %s", strings.TrimSpace(string(output)))
	}
	return nil
}

func verifyIdentityPolicy(sig sign.Signature, policy SignerPolicy) error {
	if policy.OIDCIssuer != "" && sig.OIDCIssuer != policy.OIDCIssuer {
		return fmt.Errorf("oidc issuer mismatch: got %s", sig.OIDCIssuer)
	}
	if policy.IdentityRegex != "" {
		re, err := regexp.Compile(policy.IdentityRegex)
		if err != nil {
			return fmt.Errorf("invalid identity regex: %w", err)
		}
		if !re.MatchString(sig.OIDCIdentity) {
			return fmt.Errorf("oidc identity mismatch: %s", sig.OIDCIdentity)
		}
	}
	return nil
}

func parsePublicKey(rawPEM string) (ed25519.PublicKey, error) {
	block, _ := pem.Decode([]byte(strings.TrimSpace(rawPEM)))
	if block == nil {
		return nil, fmt.Errorf("invalid public key pem")
	}
	pubAny, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}
	pub, ok := pubAny.(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("unsupported public key type")
	}
	return pub, nil
}
