package store

import "fmt"

func PublishOCI(_ string, _ string) error {
	return fmt.Errorf("oci publish is not implemented in MVP; use local store")
}
