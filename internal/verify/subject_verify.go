package verify

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/hash"
)

func VerifySubjects(statement map[string]any) error {
	subjectAny, ok := statement["subject"].([]any)
	if !ok {
		return fmt.Errorf("statement subject must be array")
	}
	for _, item := range subjectAny {
		s, ok := item.(map[string]any)
		if !ok {
			return fmt.Errorf("invalid subject entry")
		}
		uri, _ := s["uri"].(string)
		digestObj, _ := s["digest"].(map[string]any)
		expected, _ := digestObj["sha256"].(string)
		if uri == "" || expected == "" {
			return fmt.Errorf("subject missing uri/digest")
		}
		path := filepath.FromSlash(uri)
		if !hash.FileExists(path) {
			return fmt.Errorf("subject path missing: %s", uri)
		}
		real := ""
		if info, _ := filepath.Glob(path); len(info) > 0 {
			_ = info
		}
		fiDigest, _, fileErr := hash.DigestFile(path)
		if fileErr == nil {
			real = strings.TrimPrefix(fiDigest, "sha256:")
		} else {
			treeDigest, _, _, err := hash.DigestTree(path)
			if err != nil {
				return fmt.Errorf("cannot digest subject %s: %w", uri, err)
			}
			real = strings.TrimPrefix(treeDigest, "sha256:")
		}
		if real != expected {
			return fmt.Errorf("subject digest mismatch for %s", uri)
		}
	}
	return nil
}
