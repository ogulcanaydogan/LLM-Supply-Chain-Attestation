package verify

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
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
