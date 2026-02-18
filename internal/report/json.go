package report

import (
	"encoding/json"
	"os"

	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/verify"
)

func WriteJSON(path string, r verify.Report) error {
	raw, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}
