package sign

import (
	"strings"
	"testing"
)

func TestKMSSignerSignNotImplemented(t *testing.T) {
	signer := &KMSSigner{}
	_, err := signer.Sign([]byte("payload"))
	if err == nil {
		t.Fatal("expected not implemented error")
	}
	if !strings.Contains(err.Error(), "not implemented") {
		t.Fatalf("unexpected error: %v", err)
	}
}
