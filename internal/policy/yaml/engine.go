package yaml

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/sign"
	goyaml "gopkg.in/yaml.v3"
)

type Policy struct {
	Version            string   `yaml:"version"`
	OIDCIssuer         string   `yaml:"oidc_issuer"`
	IdentityRegex      string   `yaml:"identity_regex"`
	PlaintextAllowlist []string `yaml:"plaintext_allowlist"`
	Gates              []Gate   `yaml:"gates"`
}

type Gate struct {
	ID                   string   `yaml:"id"`
	TriggerPaths         []string `yaml:"trigger_paths"`
	RequiredAttestations []string `yaml:"required_attestations"`
	Message              string   `yaml:"message"`
}

type StatementView struct {
	AttestationType string
	StatementID     string
	PrivacyMode     string
}

func LoadPolicy(path string) (Policy, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Policy{}, err
	}
	var p Policy
	if err := goyaml.Unmarshal(raw, &p); err != nil {
		return Policy{}, err
	}
	return p, nil
}

func LoadStatements(source string) ([]StatementView, error) {
	fi, err := os.Stat(source)
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0)
	if fi.IsDir() {
		entries, err := os.ReadDir(source)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if strings.HasSuffix(name, ".bundle.json") || strings.HasSuffix(name, ".json") {
				paths = append(paths, filepath.Join(source, name))
			}
		}
	} else {
		paths = append(paths, source)
	}
	sort.Strings(paths)

	out := make([]StatementView, 0)
	for _, p := range paths {
		if strings.HasSuffix(p, ".bundle.json") {
			bundle, err := sign.ReadBundle(p)
			if err != nil {
				return nil, err
			}
			var payload map[string]any
			if err := sign.DecodePayload(bundle, &payload); err != nil {
				return nil, err
			}
			out = append(out, extract(payload))
			continue
		}
		raw, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		var payload map[string]any
		if err := goyaml.Unmarshal(raw, &payload); err != nil {
			return nil, err
		}
		if _, ok := payload["attestation_type"]; ok {
			out = append(out, extract(payload))
		}
	}
	return out, nil
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

func Evaluate(policy Policy, statements []StatementView, gitRef string) ([]string, error) {
	changed, err := changedFiles(gitRef)
	if err != nil {
		return nil, err
	}
	present := make(map[string]struct{})
	allowPlain := make(map[string]struct{})
	for _, id := range policy.PlaintextAllowlist {
		allowPlain[id] = struct{}{}
	}
	for _, st := range statements {
		present[st.AttestationType] = struct{}{}
		if st.PrivacyMode == "plaintext_explicit" {
			if _, ok := allowPlain[st.StatementID]; !ok {
				return []string{"Sensitive payload exposure blocked by policy."}, nil
			}
		}
	}

	violations := make([]string, 0)
	for _, gate := range policy.Gates {
		if !triggered(changed, gate.TriggerPaths) {
			continue
		}
		missing := make([]string, 0)
		for _, req := range gate.RequiredAttestations {
			if _, ok := present[req]; !ok {
				missing = append(missing, req)
			}
		}
		if len(missing) > 0 {
			msg := gate.Message
			if msg == "" {
				msg = fmt.Sprintf("%s missing attestations: %s", gate.ID, strings.Join(missing, ", "))
			}
			violations = append(violations, msg)
		}
	}
	return violations, nil
}

func triggered(changed []string, patterns []string) bool {
	for _, p := range patterns {
		for _, c := range changed {
			if match(c, p) {
				return true
			}
		}
	}
	return false
}

func match(path, pattern string) bool {
	pattern = filepath.ToSlash(pattern)
	if strings.HasSuffix(pattern, "/**") {
		prefix := strings.TrimSuffix(pattern, "/**")
		return strings.HasPrefix(path, prefix+"/") || path == prefix
	}
	ok, _ := filepath.Match(pattern, path)
	return ok
}

func extract(payload map[string]any) StatementView {
	privacyMode := ""
	if p, ok := payload["privacy"].(map[string]any); ok {
		if m, ok := p["mode"].(string); ok {
			privacyMode = m
		}
	}
	return StatementView{
		AttestationType: asString(payload["attestation_type"]),
		StatementID:     asString(payload["statement_id"]),
		PrivacyMode:     privacyMode,
	}
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}
