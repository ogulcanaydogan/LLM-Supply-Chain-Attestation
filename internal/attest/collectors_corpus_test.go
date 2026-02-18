package attest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ogulcanaydogan/llm-supply-chain-attestation/pkg/types"
)

func TestCollectCorpus(t *testing.T) {
	st, err := CollectCorpus("../../examples/tiny-rag/configs/corpus.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if st.AttestationType != types.AttestationCorpus {
		t.Fatalf("type = %q, want corpus_attestation", st.AttestationType)
	}
	if len(st.Subject) == 0 {
		t.Fatal("expected subjects")
	}
	pred, ok := st.Predicate.(types.CorpusPredicate)
	if !ok {
		t.Fatal("predicate is not CorpusPredicate")
	}
	if pred.CorpusSnapshotID == "" {
		t.Error("corpus_snapshot_id is empty")
	}
	if pred.DocumentManifestDigest == "" {
		t.Error("document_manifest_digest is empty")
	}
	if pred.ChunkingConfigDigest == "" {
		t.Error("chunking_config_digest is empty")
	}
	if pred.EmbeddingInputDigest == "" {
		t.Error("embedding_input_digest is empty")
	}
	if pred.VectorIndexDigest == "" {
		t.Error("vector_index_digest is empty")
	}
	if pred.EmbeddingModel == "" {
		t.Error("embedding_model is empty")
	}
	if len(pred.ConnectorConfigDigests) == 0 {
		t.Error("expected connector config digests")
	}
}

func TestCollectCorpus_NoDependsOn(t *testing.T) {
	st, err := CollectCorpus("../../examples/tiny-rag/configs/corpus.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if st.Annotations != nil && st.Annotations["depends_on"] != "" {
		t.Errorf("corpus should not have depends_on, got %q", st.Annotations["depends_on"])
	}
}

func TestCollectCorpus_MissingSnapshotID(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "corpus.yaml")
	content := `embedding_model: text-embedding-3-large
index_builder_image_digest: sha256:placeholder
document_manifest: doc.json
chunking_config: chunk.yaml
embedding_input: embed.jsonl
vector_index: vec.bin
`
	os.WriteFile(cfg, []byte(content), 0o644)
	for _, f := range []string{"doc.json", "chunk.yaml", "embed.jsonl", "vec.bin"} {
		os.WriteFile(filepath.Join(dir, f), []byte("test"), 0o644)
	}

	_, err := CollectCorpus(cfg)
	if err == nil {
		t.Fatal("expected error for missing corpus_snapshot_id")
	}
	if !strings.Contains(err.Error(), "corpus_snapshot_id") {
		t.Errorf("error = %q", err)
	}
}

func TestCollectCorpus_MissingEmbeddingModel(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "corpus.yaml")
	content := `corpus_snapshot_id: kb-test
index_builder_image_digest: sha256:placeholder
document_manifest: doc.json
chunking_config: chunk.yaml
embedding_input: embed.jsonl
vector_index: vec.bin
`
	os.WriteFile(cfg, []byte(content), 0o644)
	for _, f := range []string{"doc.json", "chunk.yaml", "embed.jsonl", "vec.bin"} {
		os.WriteFile(filepath.Join(dir, f), []byte("test"), 0o644)
	}

	_, err := CollectCorpus(cfg)
	if err == nil {
		t.Fatal("expected error for missing embedding_model")
	}
	if !strings.Contains(err.Error(), "embedding_model") {
		t.Errorf("error = %q", err)
	}
}

func TestCollectCorpus_MissingIndexDigest(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "corpus.yaml")
	content := `corpus_snapshot_id: kb-test
embedding_model: text-embedding-3-large
document_manifest: doc.json
chunking_config: chunk.yaml
embedding_input: embed.jsonl
vector_index: vec.bin
`
	os.WriteFile(cfg, []byte(content), 0o644)
	for _, f := range []string{"doc.json", "chunk.yaml", "embed.jsonl", "vec.bin"} {
		os.WriteFile(filepath.Join(dir, f), []byte("test"), 0o644)
	}

	_, err := CollectCorpus(cfg)
	if err == nil {
		t.Fatal("expected error for missing index_builder_image_digest")
	}
	if !strings.Contains(err.Error(), "index_builder_image_digest") {
		t.Errorf("error = %q", err)
	}
}

func TestCollectCorpus_MissingRequiredFile(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "corpus.yaml")
	content := `corpus_snapshot_id: kb-test
embedding_model: text-embedding-3-large
index_builder_image_digest: sha256:placeholder
document_manifest: nonexistent.json
chunking_config: chunk.yaml
embedding_input: embed.jsonl
vector_index: vec.bin
`
	os.WriteFile(cfg, []byte(content), 0o644)
	os.WriteFile(filepath.Join(dir, "chunk.yaml"), []byte("test"), 0o644)
	os.WriteFile(filepath.Join(dir, "embed.jsonl"), []byte("test"), 0o644)
	os.WriteFile(filepath.Join(dir, "vec.bin"), []byte("test"), 0o644)

	_, err := CollectCorpus(cfg)
	if err == nil {
		t.Fatal("expected error for missing document_manifest file")
	}
}
