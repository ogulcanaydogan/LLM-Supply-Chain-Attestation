package attest

import (
	"fmt"
	"os"
	"strings"

	"filippo.io/age"
	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/hash"
	"github.com/ogulcanaydogan/llm-supply-chain-attestation/pkg/types"
)

type privacyConfig struct {
	PrivacyMode                    string `yaml:"privacy_mode"`
	EncryptedPayloadPath           string `yaml:"encrypted_payload_path"`
	AgeRecipient                   string `yaml:"age_recipient"`
	EncryptionRecipientFingerprint string `yaml:"encryption_recipient_fingerprint"`
}

func applyPrivacyConfig(statement *types.Statement, configPath string) error {
	cfg := privacyConfig{}
	if err := LoadConfig(configPath, &cfg); err != nil {
		return err
	}

	mode := strings.TrimSpace(cfg.PrivacyMode)
	if mode == "" {
		mode = "hash_only"
	}

	switch mode {
	case "hash_only":
		statement.Privacy = types.Privacy{Mode: "hash_only"}
		return nil
	case "plaintext_explicit":
		statement.Privacy = types.Privacy{Mode: "plaintext_explicit"}
		return nil
	case "encrypted_payload":
		payloadPath := resolvePath(configPath, cfg.EncryptedPayloadPath)
		if payloadPath == "" {
			return fmt.Errorf("encrypted_payload requires encrypted_payload_path in collector config")
		}
		if cfg.AgeRecipient == "" {
			return fmt.Errorf("encrypted_payload requires age_recipient in collector config")
		}
		if _, err := age.ParseX25519Recipient(cfg.AgeRecipient); err != nil {
			return fmt.Errorf("parse age_recipient: %w", err)
		}
		raw, err := os.ReadFile(payloadPath)
		if err != nil {
			return fmt.Errorf("read encrypted payload source %s: %w", payloadPath, err)
		}

		// The statement stores only metadata. Digest is deterministically bound to
		// source bytes and age recipient material, never plaintext content.
		scope := append([]byte("age:x25519:"+cfg.AgeRecipient+"\n"), raw...)
		blobDigest := hash.DigestBytes(scope)

		fp := strings.TrimSpace(cfg.EncryptionRecipientFingerprint)
		if fp == "" {
			fp = strings.TrimPrefix(hash.DigestBytes([]byte(cfg.AgeRecipient)), "sha256:")
		}
		statement.Privacy = types.Privacy{
			Mode:                           "encrypted_payload",
			EncryptedBlobDigest:            blobDigest,
			EncryptionRecipientFingerprint: fp,
		}
		return nil
	default:
		return fmt.Errorf("unsupported privacy_mode %q", mode)
	}
}
