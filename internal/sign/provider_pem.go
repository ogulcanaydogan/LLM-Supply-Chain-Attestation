package sign

import (
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
)

type PEMSigner struct {
	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
}

func NewPEMSigner(keyPath string) (*PEMSigner, error) {
	raw, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("read pem key: %w", err)
	}
	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, fmt.Errorf("invalid pem key")
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse pkcs8 key: %w", err)
	}
	priv, ok := parsed.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("unsupported key type: need ed25519")
	}
	pub := priv.Public().(ed25519.PublicKey)
	return &PEMSigner{PrivateKey: priv, PublicKey: pub}, nil
}

func (s *PEMSigner) Sign(canonicalPayload []byte) (SignMaterial, error) {
	sig := ed25519.Sign(s.PrivateKey, canonicalPayload)
	pubPEM, err := encodePublicKeyPEM(s.PublicKey)
	if err != nil {
		return SignMaterial{}, err
	}
	h := sha256.Sum256(s.PublicKey)
	return SignMaterial{
		KeyID:        hex.EncodeToString(h[:8]),
		SigB64:       base64.StdEncoding.EncodeToString(sig),
		Provider:     "pem",
		PublicKeyPEM: pubPEM,
	}, nil
}

func encodePublicKeyPEM(pub ed25519.PublicKey) (string, error) {
	pkix, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return "", err
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pkix})), nil
}

func GeneratePEMPrivateKey(path string) error {
	_, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return err
	}
	pkcs8, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return err
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8})
	return os.WriteFile(path, pemBytes, 0o600)
}
