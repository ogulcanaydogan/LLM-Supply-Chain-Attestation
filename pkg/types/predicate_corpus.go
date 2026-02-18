package types

type NamedDigest struct {
	Name   string `json:"name"`
	Digest string `json:"digest"`
}

type CorpusPredicate struct {
	CorpusSnapshotID        string        `json:"corpus_snapshot_id"`
	ConnectorConfigDigests  []NamedDigest `json:"connector_config_digests"`
	DocumentManifestDigest  string        `json:"document_manifest_digest"`
	ChunkingConfigDigest    string        `json:"chunking_config_digest"`
	EmbeddingModel          string        `json:"embedding_model"`
	EmbeddingInputDigest    string        `json:"embedding_input_digest"`
	IndexBuilderImageDigest string        `json:"index_builder_image_digest"`
	VectorIndexDigest       string        `json:"vector_index_digest"`
	BuildCommandDigest      string        `json:"build_command_digest,omitempty"`
}
