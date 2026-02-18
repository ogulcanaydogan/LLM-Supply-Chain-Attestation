package attest

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/hash"
	"github.com/ogulcanaydogan/llm-supply-chain-attestation/pkg/types"
)

func newStatement(attType string, predicate any, subjects []types.Subject, materials []types.Subject) types.Statement {
	return types.Statement{
		SchemaVersion:   "1.0.0",
		StatementID:     uuid.NewString(),
		AttestationType: attType,
		PredicateType:   types.PredicateURI(attType),
		GeneratedAt:     time.Now().UTC().Format(time.RFC3339),
		Generator: types.Generator{
			Name:    "llmsa",
			Version: "0.1.0",
			GitSHA:  readGitSHA(),
		},
		Subject:   subjects,
		Materials: materials,
		Predicate: predicate,
		Privacy: types.Privacy{
			Mode: "hash_only",
		},
		Annotations: map[string]string{
			"generated_by": "llmsa attest create",
		},
	}
}

func setDependsOn(statement *types.Statement, deps ...string) {
	if statement == nil {
		return
	}
	unique := make(map[string]struct{})
	for _, dep := range deps {
		dep = strings.TrimSpace(dep)
		if dep == "" {
			continue
		}
		unique[dep] = struct{}{}
	}
	if len(unique) == 0 {
		return
	}
	items := make([]string, 0, len(unique))
	for dep := range unique {
		items = append(items, dep)
	}
	sort.Strings(items)
	if statement.Annotations == nil {
		statement.Annotations = map[string]string{}
	}
	statement.Annotations["depends_on"] = strings.Join(items, ",")
}

func readGitSHA() string {
	if v := os.Getenv("GITHUB_SHA"); v != "" {
		return v
	}
	return "local"
}

func subjectFromPath(path string) (types.Subject, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return types.Subject{}, err
	}
	if fi.IsDir() {
		digest, _, _, err := hash.DigestTree(path)
		if err != nil {
			return types.Subject{}, err
		}
		return types.Subject{
			Name:      filepath.Base(path),
			URI:       filepath.ToSlash(path),
			Digest:    types.Digest{SHA256: strings.TrimPrefix(digest, "sha256:")},
			SizeBytes: 0,
		}, nil
	}
	digest, size, err := hash.DigestFile(path)
	if err != nil {
		return types.Subject{}, err
	}
	return types.Subject{
		Name:      filepath.Base(path),
		URI:       filepath.ToSlash(path),
		Digest:    types.Digest{SHA256: strings.TrimPrefix(digest, "sha256:")},
		SizeBytes: size,
	}, nil
}

func sortedFileDigests(dir string) ([]string, []types.Subject, error) {
	d, _, entries, err := hash.DigestTree(dir)
	if err != nil {
		return nil, nil, err
	}
	_ = d
	digests := make([]string, 0, len(entries))
	subjects := make([]types.Subject, 0, len(entries))
	for _, e := range entries {
		digests = append(digests, e.Digest)
		subjects = append(subjects, types.Subject{
			Name:      e.Path,
			URI:       filepath.ToSlash(filepath.Join(dir, e.Path)),
			Digest:    types.Digest{SHA256: strings.TrimPrefix(e.Digest, "sha256:")},
			SizeBytes: e.Size,
		})
	}
	sort.Strings(digests)
	return digests, subjects, nil
}

func digestOfString(value string) string {
	return hash.DigestBytes([]byte(value))
}

func bundleDigest(parts ...string) string {
	sort.Strings(parts)
	return hash.DigestBytes([]byte(strings.Join(parts, "\n")))
}

func requirePath(path string, name string) error {
	if path == "" {
		return fmt.Errorf("%s path is required", name)
	}
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("%s path %s: %w", name, path, err)
	}
	return nil
}

func resolvePath(configPath, candidate string) string {
	if candidate == "" {
		return candidate
	}
	if filepath.IsAbs(candidate) {
		return candidate
	}
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}
	cfgDir := filepath.Dir(configPath)
	joined := filepath.Clean(filepath.Join(cfgDir, candidate))
	if _, err := os.Stat(joined); err == nil {
		return joined
	}
	return candidate
}
