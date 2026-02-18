package verify

import (
	"fmt"
	"time"
)

func VerifyBasicChainConstraints(statement map[string]any) error {
	generatedAt, _ := statement["generated_at"].(string)
	if generatedAt == "" {
		return fmt.Errorf("generated_at is required")
	}
	if _, err := time.Parse(time.RFC3339, generatedAt); err != nil {
		return fmt.Errorf("invalid generated_at: %w", err)
	}
	return nil
}
