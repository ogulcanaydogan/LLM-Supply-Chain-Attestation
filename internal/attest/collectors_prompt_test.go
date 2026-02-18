package attest

import (
	"testing"

	"github.com/ogulcanaydogan/llm-supply-chain-attestation/pkg/types"
)

func TestCollectPrompt(t *testing.T) {
	st, err := CollectPrompt("../../examples/tiny-rag/configs/prompt.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if st.AttestationType != types.AttestationPrompt {
		t.Fatalf("unexpected attestation type: %s", st.AttestationType)
	}
	if len(st.Subject) == 0 {
		t.Fatalf("expected subjects")
	}
}
