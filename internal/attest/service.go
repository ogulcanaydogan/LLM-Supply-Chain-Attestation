package attest

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/hash"
	"github.com/ogulcanaydogan/llm-supply-chain-attestation/pkg/types"
)

type CreateOptions struct {
	Type             string
	ConfigPath       string
	OutDir           string
	DeterminismCheck int
}

func CreateByType(opts CreateOptions) ([]string, error) {
	if opts.OutDir == "" {
		return nil, fmt.Errorf("--out is required")
	}
	if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
		return nil, fmt.Errorf("create out dir: %w", err)
	}
	statement, err := collectByType(opts.Type, opts.ConfigPath)
	if err != nil {
		return nil, err
	}
	if err := applyPrivacyConfig(&statement, opts.ConfigPath); err != nil {
		return nil, err
	}

	if opts.DeterminismCheck > 1 {
		first, _, err := hash.HashCanonicalJSON(statement)
		if err != nil {
			return nil, err
		}
		for i := 0; i < opts.DeterminismCheck-1; i++ {
			again, err := collectByType(opts.Type, opts.ConfigPath)
			if err != nil {
				return nil, err
			}
			// Determinism check validates content hashing, not runtime nonce fields.
			again.StatementID = statement.StatementID
			again.GeneratedAt = statement.GeneratedAt
			next, _, err := hash.HashCanonicalJSON(again)
			if err != nil {
				return nil, err
			}
			if first != next {
				return nil, fmt.Errorf("determinism check failed: %s != %s", first, next)
			}
		}
	}

	fileName := fmt.Sprintf("statement_%s_%s.json", statement.AttestationType, statement.StatementID)
	outPath := filepath.Join(opts.OutDir, fileName)
	raw, err := json.MarshalIndent(statement, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal statement: %w", err)
	}
	if err := os.WriteFile(outPath, raw, 0o644); err != nil {
		return nil, fmt.Errorf("write statement: %w", err)
	}
	return []string{outPath}, nil
}

func CreateChangedOnly(gitRef, outDir string, determinismCheck int) ([]string, error) {
	cfg := DefaultProjectConfig()
	if hash.FileExists("llmsa.yaml") {
		if err := LoadConfig("llmsa.yaml", &cfg); err != nil {
			return nil, err
		}
	}
	changed, err := changedFiles(gitRef)
	if err != nil {
		return nil, err
	}
	typesToCreate := inferAttestationTypes(changed, cfg.PathRules)
	if len(typesToCreate) == 0 {
		return nil, fmt.Errorf("no changed artifacts mapped to attestation rules")
	}
	sort.Strings(typesToCreate)
	created := make([]string, 0)
	for _, attType := range typesToCreate {
		cfgPath := cfg.Collectors[attType]
		if cfgPath == "" {
			return nil, fmt.Errorf("missing collector config for %s", attType)
		}
		out, err := CreateByType(CreateOptions{Type: attType, ConfigPath: cfgPath, OutDir: outDir, DeterminismCheck: determinismCheck})
		if err != nil {
			return nil, err
		}
		created = append(created, out...)
	}
	return created, nil
}

func collectByType(attType, configPath string) (types.Statement, error) {
	switch attType {
	case types.AttestationPrompt:
		return CollectPrompt(configPath)
	case types.AttestationCorpus:
		return CollectCorpus(configPath)
	case types.AttestationEval:
		return CollectEval(configPath)
	case types.AttestationRoute:
		return CollectRoute(configPath)
	case types.AttestationSLO:
		return CollectSLO(configPath)
	default:
		return types.Statement{}, fmt.Errorf("unsupported attestation type: %s", attType)
	}
}

func changedFiles(gitRef string) ([]string, error) {
	if gitRef == "" {
		gitRef = "HEAD~1"
	}
	if err := exec.Command("git", "rev-parse", "--verify", "HEAD").Run(); err != nil {
		return []string{}, nil
	}
	cmd := exec.Command("git", "diff", "--name-only", gitRef+"...HEAD")
	raw, err := cmd.Output()
	if err != nil {
		return []string{}, nil
	}
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	out := make([]string, 0)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		out = append(out, filepath.ToSlash(line))
	}
	return out, nil
}

func inferAttestationTypes(changed []string, rules map[string][]string) []string {
	seen := make(map[string]struct{})
	for _, path := range changed {
		for attType, patterns := range rules {
			for _, p := range patterns {
				if matches(path, p) {
					seen[attType] = struct{}{}
					break
				}
			}
		}
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	return out
}

func matches(path, pattern string) bool {
	pattern = filepath.ToSlash(pattern)
	if strings.HasSuffix(pattern, "/**") {
		prefix := strings.TrimSuffix(pattern, "/**")
		return strings.HasPrefix(path, prefix+"/") || path == prefix
	}
	ok, _ := filepath.Match(pattern, path)
	if ok {
		return true
	}
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(path, strings.TrimSuffix(pattern, "*"))
	}
	return false
}
