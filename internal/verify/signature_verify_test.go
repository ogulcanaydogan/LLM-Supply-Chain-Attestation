package verify

import (
	"strings"
	"testing"

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
