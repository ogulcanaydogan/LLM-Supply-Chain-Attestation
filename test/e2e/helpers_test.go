//go:build e2e

package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/attest"
	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/hash"
	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/sign"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("cannot resolve test file path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
}

func schemaDir(t *testing.T) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "schemas", "v1")
}

func configPath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "examples", "tiny-rag", "configs", name+".yaml")
}

func createAndSignAttestation(t *testing.T, attType, cfgPath, outDir, keyPath string) string {
	t.Helper()
	files, err := attest.CreateByType(attest.CreateOptions{
		Type:       attType,
		ConfigPath: cfgPath,
		OutDir:     outDir,
	})
	if err != nil {
		t.Fatalf("create %s: %v", attType, err)
	}
	if len(files) == 0 {
		t.Fatalf("no files for %s", attType)
	}
	raw, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatal(err)
	}
	var statement map[string]any
	if err := json.Unmarshal(raw, &statement); err != nil {
		t.Fatal(err)
	}
	canonical, err := hash.CanonicalJSON(statement)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := sign.NewPEMSigner(keyPath)
	if err != nil {
		t.Fatal(err)
	}
	material, err := signer.Sign(canonical)
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := sign.CreateBundle(statement, material)
	if err != nil {
		t.Fatal(err)
	}
	bundlePath := filepath.Join(outDir, attType+".bundle.json")
	if err := sign.WriteBundle(bundlePath, bundle); err != nil {
		t.Fatal(err)
	}
	return bundlePath
}

func createAllFiveAttestations(t *testing.T, outDir, keyPath string) []string {
	t.Helper()
	types := []struct{ attType, cfg string }{
		{"prompt_attestation", "prompt"},
		{"corpus_attestation", "corpus"},
		{"eval_attestation", "eval"},
		{"route_attestation", "route"},
		{"slo_attestation", "slo"},
	}
	var paths []string
	for _, tt := range types {
		p := createAndSignAttestation(t, tt.attType, configPath(t, tt.cfg), outDir, keyPath)
		paths = append(paths, p)
	}
	return paths
}
