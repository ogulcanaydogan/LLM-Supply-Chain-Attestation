package rego

import "fmt"

func Evaluate(_ string, _ any) error {
	return fmt.Errorf("rego engine not enabled in MVP; use YAML policy engine")
}
