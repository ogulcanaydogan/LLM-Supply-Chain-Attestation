package sign

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
)

type SigstoreSigner struct {
	PEMKeyPath string
	Issuer     string
	Identity   string
}

func (s *SigstoreSigner) Sign(canonicalPayload []byte) (SignMaterial, error) {
	issuer := s.Issuer
	if issuer == "" {
		issuer = "https://token.actions.githubusercontent.com"
	}
	identity := s.Identity
	if identity == "" {
		repo := os.Getenv("GITHUB_REPOSITORY")
		wf := os.Getenv("GITHUB_WORKFLOW")
		if repo == "" {
			repo = "local/dev"
		}
		if wf == "" {
			wf = "manual"
		}
		identity = fmt.Sprintf("repo:%s:workflow:%s", repo, wf)
	}

	var priv ed25519.PrivateKey
	var pub ed25519.PublicKey
	if s.PEMKeyPath != "" {
		pemSigner, err := NewPEMSigner(s.PEMKeyPath)
		if err != nil {
			return SignMaterial{}, err
		}
		priv = pemSigner.PrivateKey
		pub = pemSigner.PublicKey
	} else {
		generatedPub, generatedPriv, err := ed25519.GenerateKey(nil)
		if err != nil {
			return SignMaterial{}, err
		}
		pub = generatedPub
		priv = generatedPriv
	}

	sig := ed25519.Sign(priv, canonicalPayload)
	pubPEM, err := encodePublicKeyPEM(pub)
	if err != nil {
		return SignMaterial{}, err
	}
	h := sha256.Sum256(pub)
	return SignMaterial{
		KeyID:        "sigstore-" + hex.EncodeToString(h[:6]),
		SigB64:       base64.StdEncoding.EncodeToString(sig),
		Provider:     "sigstore",
		PublicKeyPEM: pubPEM,
		OIDCIssuer:   issuer,
		OIDCIdentity: identity,
	}, nil
}
