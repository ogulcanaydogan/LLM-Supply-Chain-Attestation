package attest

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type ProjectConfig struct {
	Collectors map[string]string   `yaml:"collectors"`
	PathRules  map[string][]string `yaml:"path_rules"`
}

func LoadConfig(path string, out any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config %s: %w", path, err)
	}
	if err := yaml.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("parse config %s: %w", path, err)
	}
	return nil
}

func DefaultProjectConfig() ProjectConfig {
	return ProjectConfig{
		Collectors: map[string]string{
			"prompt_attestation": "examples/tiny-rag/configs/prompt.yaml",
			"corpus_attestation": "examples/tiny-rag/configs/corpus.yaml",
			"eval_attestation":   "examples/tiny-rag/configs/eval.yaml",
			"route_attestation":  "examples/tiny-rag/configs/route.yaml",
			"slo_attestation":    "examples/tiny-rag/configs/slo.yaml",
		},
		PathRules: map[string][]string{
			"prompt_attestation": {"prompt/**", "prompts/**", "examples/tiny-rag/app/**"},
			"corpus_attestation": {"corpus/**", "data/**", "examples/tiny-rag/data/**"},
			"eval_attestation":   {"eval/**", "examples/tiny-rag/eval/**"},
			"route_attestation":  {"route/**", "examples/tiny-rag/route/**"},
			"slo_attestation":    {"slo/**", "examples/tiny-rag/slo/**"},
		},
	}
}
