package attest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ogulcanaydogan/llm-supply-chain-attestation/pkg/types"
)

// --- matches() ---

func TestMatchesDirectoryRecursive(t *testing.T) {
	if !matches("corpus/data/file.txt", "corpus/**") {
		t.Error("expected corpus/data/file.txt to match corpus/**")
	}
}

func TestMatchesDirectoryRecursiveExactRoot(t *testing.T) {
	if !matches("corpus", "corpus/**") {
		t.Error("expected corpus to match corpus/**")
	}
}

func TestMatchesDirectoryRecursiveNoMatch(t *testing.T) {
	if matches("other/file.txt", "corpus/**") {
		t.Error("expected other/file.txt NOT to match corpus/**")
	}
}

func TestMatchesStandardGlob(t *testing.T) {
	if !matches("prompt.yaml", "*.yaml") {
		t.Error("expected prompt.yaml to match *.yaml")
	}
}

func TestMatchesTrailingWildcard(t *testing.T) {
	if !matches("promptfoo.yaml", "prompt*") {
		t.Error("expected promptfoo.yaml to match prompt*")
	}
}

func TestMatchesNoMatch(t *testing.T) {
	if matches("unrelated.go", "data/**") {
		t.Error("expected unrelated.go NOT to match data/**")
	}
}

// --- inferAttestationTypes() ---

func TestInferAttestationTypesSingleMatch(t *testing.T) {
	rules := map[string][]string{
		"prompt_attestation": {"prompt/**"},
	}
	got := inferAttestationTypes([]string{"prompt/system.txt"}, rules)
	if len(got) != 1 || got[0] != "prompt_attestation" {
		t.Fatalf("expected [prompt_attestation], got %v", got)
	}
}

func TestInferAttestationTypesMultipleMatches(t *testing.T) {
	rules := map[string][]string{
		"prompt_attestation": {"prompt/**"},
		"corpus_attestation": {"corpus/**"},
	}
	got := inferAttestationTypes([]string{"prompt/sys.txt", "corpus/data.csv"}, rules)
	if len(got) != 2 {
		t.Fatalf("expected 2 types, got %d: %v", len(got), got)
	}
}

func TestInferAttestationTypesNoMatches(t *testing.T) {
	rules := map[string][]string{
		"prompt_attestation": {"prompt/**"},
	}
	got := inferAttestationTypes([]string{"unrelated/file.go"}, rules)
	if len(got) != 0 {
		t.Fatalf("expected 0 types, got %d: %v", len(got), got)
	}
}

func TestInferAttestationTypesDedup(t *testing.T) {
	rules := map[string][]string{
		"prompt_attestation": {"prompt/**"},
	}
	got := inferAttestationTypes([]string{"prompt/a.txt", "prompt/b.txt"}, rules)
	if len(got) != 1 {
		t.Fatalf("expected 1 type after dedup, got %d: %v", len(got), got)
	}
}

// --- collectByType() ---

func TestCollectByTypeUnsupported(t *testing.T) {
	_, err := collectByType("nonexistent_type", "/dev/null")
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
	if !strings.Contains(err.Error(), "unsupported attestation type") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- DefaultProjectConfig() ---

func TestDefaultProjectConfigStructure(t *testing.T) {
	cfg := DefaultProjectConfig()
	if len(cfg.Collectors) != 5 {
		t.Fatalf("expected 5 collectors, got %d", len(cfg.Collectors))
	}
	if len(cfg.PathRules) != 5 {
		t.Fatalf("expected 5 path rules, got %d", len(cfg.PathRules))
	}
	for _, k := range []string{"prompt_attestation", "corpus_attestation", "eval_attestation", "route_attestation", "slo_attestation"} {
		if cfg.Collectors[k] == "" {
			t.Errorf("missing collector for %s", k)
		}
		if len(cfg.PathRules[k]) == 0 {
			t.Errorf("missing path rules for %s", k)
		}
	}
}

// --- LoadConfig() ---

func TestLoadConfigFileNotFound(t *testing.T) {
	var out map[string]any
	err := LoadConfig("/nonexistent/file.yaml", &out)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "read config") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "bad.yaml")
	// Write invalid YAML that will cause unmarshal error when parsed into a struct
	os.WriteFile(path, []byte("valid: yaml"), 0o644)

	// Try loading into a string which should fail type assertion
	var out string
	err := LoadConfig(path, &out)
	if err == nil {
		t.Fatal("expected error for invalid yaml target")
	}
	if !strings.Contains(err.Error(), "parse config") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadConfigValid(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "cfg.yaml")
	os.WriteFile(path, []byte("collectors:\n  prompt_attestation: configs/prompt.yaml\n"), 0o644)

	var cfg ProjectConfig
	if err := LoadConfig(path, &cfg); err != nil {
		t.Fatalf("expected success: %v", err)
	}
	if cfg.Collectors["prompt_attestation"] != "configs/prompt.yaml" {
		t.Fatalf("unexpected collector: %q", cfg.Collectors["prompt_attestation"])
	}
}

// --- readGitSHA() ---

func TestReadGitSHAFromEnv(t *testing.T) {
	t.Setenv("GITHUB_SHA", "abc123def")
	got := readGitSHA()
	if got != "abc123def" {
		t.Fatalf("expected abc123def, got %q", got)
	}
}

func TestReadGitSHADefault(t *testing.T) {
	t.Setenv("GITHUB_SHA", "")
	got := readGitSHA()
	if got != "local" {
		t.Fatalf("expected local, got %q", got)
	}
}

// --- setDependsOn() ---

func TestSetDependsOnNilStatement(t *testing.T) {
	// Should not panic
	setDependsOn(nil, "dep1")
}

func TestSetDependsOnAllEmptyDeps(t *testing.T) {
	stmt := types.Statement{Annotations: map[string]string{}}
	setDependsOn(&stmt, "", "  ", "")
	if _, ok := stmt.Annotations["depends_on"]; ok {
		t.Fatal("expected no depends_on for all-empty deps")
	}
}

func TestSetDependsOnNilAnnotations(t *testing.T) {
	stmt := types.Statement{Annotations: nil}
	setDependsOn(&stmt, "eval_attestation")
	if stmt.Annotations == nil {
		t.Fatal("expected annotations to be initialized")
	}
	if stmt.Annotations["depends_on"] != "eval_attestation" {
		t.Fatalf("unexpected depends_on: %q", stmt.Annotations["depends_on"])
	}
}

func TestSetDependsOnDeduplicates(t *testing.T) {
	stmt := types.Statement{}
	setDependsOn(&stmt, "a", "b", "a")
	got := stmt.Annotations["depends_on"]
	if got != "a,b" {
		t.Fatalf("expected deduped sorted 'a,b', got %q", got)
	}
}

// --- requirePath() ---

func TestRequirePathEmpty(t *testing.T) {
	err := requirePath("", "test_field")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
	if !strings.Contains(err.Error(), "test_field path is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRequirePathNonexistent(t *testing.T) {
	err := requirePath("/nonexistent/path/file.txt", "config")
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
	if !strings.Contains(err.Error(), "config path") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRequirePathValid(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "file.txt")
	os.WriteFile(f, []byte("data"), 0o644)
	if err := requirePath(f, "test"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- resolvePath() ---

func TestResolvePathEmpty(t *testing.T) {
	got := resolvePath("/some/config.yaml", "")
	if got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

func TestResolvePathAbsolute(t *testing.T) {
	got := resolvePath("/some/config.yaml", "/abs/path.txt")
	if got != "/abs/path.txt" {
		t.Fatalf("expected /abs/path.txt, got %q", got)
	}
}

func TestResolvePathRelativeExists(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "data.txt")
	os.WriteFile(f, []byte("x"), 0o644)
	// If the candidate exists relative to CWD, return as-is
	got := resolvePath("/some/config.yaml", f)
	if got != f {
		t.Fatalf("expected %q, got %q", f, got)
	}
}

func TestResolvePathRelativeViaConfigDir(t *testing.T) {
	tmp := t.TempDir()
	// Create a file next to the config
	cfgPath := filepath.Join(tmp, "configs", "prompt.yaml")
	os.MkdirAll(filepath.Dir(cfgPath), 0o755)
	os.WriteFile(cfgPath, []byte("x"), 0o644)

	target := filepath.Join(tmp, "configs", "secret.txt")
	os.WriteFile(target, []byte("x"), 0o644)

	got := resolvePath(cfgPath, "secret.txt")
	if got != target {
		t.Fatalf("expected %q, got %q", target, got)
	}
}

// --- subjectFromPath() ---

func TestSubjectFromPathDirectory(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "a.txt"), []byte("aaa"), 0o644)
	os.WriteFile(filepath.Join(tmp, "b.txt"), []byte("bbb"), 0o644)

	subj, err := subjectFromPath(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if subj.Digest.SHA256 == "" {
		t.Error("expected non-empty digest for directory")
	}
	if subj.SizeBytes != 0 {
		t.Errorf("expected SizeBytes=0 for directory, got %d", subj.SizeBytes)
	}
}

func TestSubjectFromPathNonexistent(t *testing.T) {
	_, err := subjectFromPath("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
}

// --- CreateByType() ---

func TestCreateByTypeEmptyOutDir(t *testing.T) {
	_, err := CreateByType(CreateOptions{Type: "prompt_attestation", ConfigPath: "/dev/null", OutDir: ""})
	if err == nil {
		t.Fatal("expected error for empty OutDir")
	}
	if !strings.Contains(err.Error(), "--out is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateByTypeUnsupportedType(t *testing.T) {
	_, err := CreateByType(CreateOptions{Type: "bogus_type", ConfigPath: "/dev/null", OutDir: t.TempDir()})
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
	if !strings.Contains(err.Error(), "unsupported attestation type") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- digestOfString() ---

func TestDigestOfString(t *testing.T) {
	d := digestOfString("hello")
	if !strings.HasPrefix(d, "sha256:") {
		t.Fatalf("expected sha256: prefix, got %q", d)
	}
	if len(d) != 7+64 { // sha256: + 64 hex chars
		t.Fatalf("unexpected digest length: %d", len(d))
	}
}

// --- bundleDigest() ---

func TestBundleDigestDeterministic(t *testing.T) {
	d1 := bundleDigest("b", "a", "c")
	d2 := bundleDigest("a", "c", "b")
	if d1 != d2 {
		t.Fatalf("bundleDigest should be order-independent: %q != %q", d1, d2)
	}
}

// --- sortedFileDigests() ---

func TestSortedFileDigests(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "z.txt"), []byte("zzz"), 0o644)
	os.WriteFile(filepath.Join(tmp, "a.txt"), []byte("aaa"), 0o644)

	digests, subjects, err := sortedFileDigests(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(digests) != 2 || len(subjects) != 2 {
		t.Fatalf("expected 2 entries each, got digests=%d subjects=%d", len(digests), len(subjects))
	}
	// Digests should be sorted
	if digests[0] > digests[1] {
		t.Error("digests should be sorted")
	}
}

// --- newStatement() ---

func TestNewStatementFields(t *testing.T) {
	pred := map[string]string{"key": "val"}
	stmt := newStatement("prompt_attestation", pred, nil, nil)
	if stmt.SchemaVersion != "1.0.0" {
		t.Errorf("unexpected schema_version: %q", stmt.SchemaVersion)
	}
	if stmt.AttestationType != "prompt_attestation" {
		t.Errorf("unexpected attestation_type: %q", stmt.AttestationType)
	}
	if stmt.StatementID == "" {
		t.Error("expected non-empty statement_id")
	}
	if stmt.GeneratedAt == "" {
		t.Error("expected non-empty generated_at")
	}
	if stmt.Privacy.Mode != "hash_only" {
		t.Errorf("expected default privacy hash_only, got %q", stmt.Privacy.Mode)
	}
	if stmt.Annotations["generated_by"] != "llmsa attest create" {
		t.Error("missing generated_by annotation")
	}
}
