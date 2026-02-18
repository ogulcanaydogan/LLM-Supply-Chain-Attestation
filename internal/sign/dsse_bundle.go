package sign

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/hash"
)

type Bundle struct {
	Envelope Envelope `json:"envelope"`
	Metadata Metadata `json:"metadata"`
}

type Envelope struct {
	PayloadType string      `json:"payloadType"`
	Payload     string      `json:"payload"`
	Signatures  []Signature `json:"signatures"`
}

type Signature struct {
	KeyID        string `json:"keyid"`
	Sig          string `json:"sig"`
	Provider     string `json:"provider"`
	PublicKeyPEM string `json:"public_key_pem"`
	OIDCIssuer   string `json:"oidc_issuer,omitempty"`
	OIDCIdentity string `json:"oidc_identity,omitempty"`
}

type Metadata struct {
	BundleVersion string `json:"bundle_version"`
	CreatedAt     string `json:"created_at"`
	StatementHash string `json:"statement_hash"`
}

type SignMaterial struct {
	KeyID        string
	SigB64       string
	Provider     string
	PublicKeyPEM string
	OIDCIssuer   string
	OIDCIdentity string
}

func CreateBundle(statement any, material SignMaterial) (Bundle, error) {
	canonical, err := hash.CanonicalJSON(statement)
	if err != nil {
		return Bundle{}, err
	}
	statementHash := hash.DigestBytes(canonical)

	bundle := Bundle{
		Envelope: Envelope{
			PayloadType: "application/vnd.llmsa.statement.v1+json",
			Payload:     base64.StdEncoding.EncodeToString(canonical),
			Signatures: []Signature{{
				KeyID:        material.KeyID,
				Sig:          material.SigB64,
				Provider:     material.Provider,
				PublicKeyPEM: material.PublicKeyPEM,
				OIDCIssuer:   material.OIDCIssuer,
				OIDCIdentity: material.OIDCIdentity,
			}},
		},
		Metadata: Metadata{
			BundleVersion: "1",
			CreatedAt:     time.Now().UTC().Format(time.RFC3339),
			StatementHash: statementHash,
		},
	}
	return bundle, nil
}

func DecodePayload(bundle Bundle, out any) error {
	raw, err := base64.StdEncoding.DecodeString(bundle.Envelope.Payload)
	if err != nil {
		return fmt.Errorf("decode bundle payload: %w", err)
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("unmarshal bundle payload: %w", err)
	}
	return nil
}

func WriteBundle(path string, b Bundle) error {
	raw, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

func ReadBundle(path string) (Bundle, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Bundle{}, err
	}
	var b Bundle
	if err := json.Unmarshal(raw, &b); err != nil {
		return Bundle{}, err
	}
	return b, nil
}
