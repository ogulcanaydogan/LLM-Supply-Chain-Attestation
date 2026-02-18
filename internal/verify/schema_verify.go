package verify

import (
	"fmt"
	"path/filepath"

	"github.com/ogulcanaydogan/llm-supply-chain-attestation/pkg/schema"
)

func VerifySchemas(schemaDir string, statement map[string]any) error {
	baseSchema := filepath.Join(schemaDir, "statement.schema.json")
	if errs, err := schema.Validate(baseSchema, statement); err != nil {
		return err
	} else if len(errs) > 0 {
		return fmt.Errorf("statement schema invalid: %v", errs)
	}

	attType, _ := statement["attestation_type"].(string)
	predicate, _ := statement["predicate"].(map[string]any)
	if attType == "" || predicate == nil {
		return fmt.Errorf("statement missing attestation_type or predicate")
	}
	predSchema := filepath.Join(schemaDir, attType+".schema.json")
	if errs, err := schema.Validate(predSchema, predicate); err != nil {
		return err
	} else if len(errs) > 0 {
		return fmt.Errorf("predicate schema invalid: %v", errs)
	}
	return nil
}
