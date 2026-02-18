package sign

import "fmt"

type KMSSigner struct{}

func (s *KMSSigner) Sign(_ []byte) (SignMaterial, error) {
	return SignMaterial{}, fmt.Errorf("kms provider is not implemented in MVP")
}
