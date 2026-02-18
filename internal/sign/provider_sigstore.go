package sign

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type SigstoreSigner struct {
	PEMKeyPath string
	Issuer     string
	Identity   string
}

func (s *SigstoreSigner) Sign(canonicalPayload []byte) (SignMaterial, error) {
	issuer, identity := defaultOIDCClaims(s.Issuer, s.Identity)

	if s.PEMKeyPath != "" {
		pemSigner, err := NewPEMSigner(s.PEMKeyPath)
		if err != nil {
			return SignMaterial{}, err
		}
		material, err := pemSigner.Sign(canonicalPayload)
		if err != nil {
			return SignMaterial{}, err
		}
		material.Provider = "sigstore"
		material.OIDCIssuer = issuer
		material.OIDCIdentity = identity
		return material, nil
	}

	if _, err := exec.LookPath("cosign"); err != nil {
		return SignMaterial{}, fmt.Errorf("cosign binary not found for keyless signing: %w", err)
	}

	tmp, err := os.MkdirTemp("", "llmsa-sigstore-sign-")
	if err != nil {
		return SignMaterial{}, err
	}
	defer os.RemoveAll(tmp)

	payloadPath := filepath.Join(tmp, "payload.json")
	sigPath := filepath.Join(tmp, "payload.sig")
	certPath := filepath.Join(tmp, "payload.pem")
	if err := os.WriteFile(payloadPath, canonicalPayload, 0o600); err != nil {
		return SignMaterial{}, err
	}

	args := []string{"sign-blob", "--yes", "--output-signature", sigPath, "--output-certificate", certPath, payloadPath}
	cmd := exec.Command("cosign", args...)
	cmd.Env = append(os.Environ(), "COSIGN_YES=true")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return SignMaterial{}, fmt.Errorf("cosign keyless sign-blob failed: %w", err)
	}

	sigRaw, err := os.ReadFile(sigPath)
	if err != nil {
		return SignMaterial{}, err
	}
	certRaw, err := os.ReadFile(certPath)
	if err != nil {
		return SignMaterial{}, err
	}
	return signMaterialFromCosignOutputs(strings.TrimSpace(string(sigRaw)), strings.TrimSpace(string(certRaw)), issuer, identity)
}

func defaultOIDCClaims(issuer, identity string) (string, string) {
	if issuer == "" {
		issuer = "https://token.actions.githubusercontent.com"
	}
	if identity == "" {
		if wfRef := os.Getenv("GITHUB_WORKFLOW_REF"); wfRef != "" {
			identity = "https://github.com/" + wfRef
		} else {
			repo := os.Getenv("GITHUB_REPOSITORY")
			wf := os.Getenv("GITHUB_WORKFLOW")
			ref := os.Getenv("GITHUB_REF")
			if repo == "" {
				repo = "local/dev"
			}
			if wf == "" {
				wf = "manual.yml"
			}
			if ref == "" {
				ref = "refs/heads/local"
			}
			identity = fmt.Sprintf("https://github.com/%s/.github/workflows/%s@%s", repo, wf, ref)
		}
	}
	return issuer, identity
}

func publicKeyFromCertificatePEM(certPEM string) (string, string, error) {
	block, _ := pem.Decode([]byte(strings.TrimSpace(certPEM)))
	if block == nil {
		return "", "", fmt.Errorf("invalid certificate pem")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", "", fmt.Errorf("parse certificate: %w", err)
	}
	pkix, err := x509.MarshalPKIXPublicKey(cert.PublicKey)
	if err != nil {
		return "", "", fmt.Errorf("marshal cert public key: %w", err)
	}
	pubPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pkix}))
	sum := sha256.Sum256(block.Bytes)
	return pubPEM, hex.EncodeToString(sum[:]), nil
}

func signMaterialFromCosignOutputs(sigText string, certText string, issuer string, identity string) (SignMaterial, error) {
	if sigText == "" {
		return SignMaterial{}, fmt.Errorf("cosign produced empty signature output")
	}
	if certText == "" {
		return SignMaterial{}, fmt.Errorf("cosign produced empty certificate output")
	}

	// Some cosign versions/issuers may emit certificate formats we cannot parse here.
	// Keep certificate bytes for verification and derive keyid from raw certificate digest.
	pubPEM := ""
	certSum := sha256.Sum256([]byte(certText))
	certDigest := hex.EncodeToString(certSum[:])
	if parsedPub, parsedDigest, parseErr := publicKeyFromCertificatePEM(certText); parseErr == nil {
		pubPEM = parsedPub
		certDigest = parsedDigest
	}

	return SignMaterial{
		KeyID:          "sigstore-" + certDigest[:12],
		SigB64:         sigText,
		Provider:       "sigstore",
		PublicKeyPEM:   pubPEM,
		CertificatePEM: certText,
		OIDCIssuer:     issuer,
		OIDCIdentity:   identity,
	}, nil
}
