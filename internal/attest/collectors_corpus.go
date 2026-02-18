package attest

import (
	"fmt"
	"path/filepath"

	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/hash"
	"github.com/ogulcanaydogan/llm-supply-chain-attestation/pkg/types"
)

type CorpusConfig struct {
	CorpusSnapshotID        string   `yaml:"corpus_snapshot_id"`
	ConnectorConfigs        []string `yaml:"connector_configs"`
	DocumentManifest        string   `yaml:"document_manifest"`
	ChunkingConfig          string   `yaml:"chunking_config"`
	EmbeddingModel          string   `yaml:"embedding_model"`
	EmbeddingInput          string   `yaml:"embedding_input"`
	IndexBuilderImageDigest string   `yaml:"index_builder_image_digest"`
	VectorIndex             string   `yaml:"vector_index"`
	BuildCommand            string   `yaml:"build_command"`
}

func CollectCorpus(configPath string) (types.Statement, error) {
	cfg := CorpusConfig{}
	if err := LoadConfig(configPath, &cfg); err != nil {
		return types.Statement{}, err
	}
	for i := range cfg.ConnectorConfigs {
		cfg.ConnectorConfigs[i] = resolvePath(configPath, cfg.ConnectorConfigs[i])
	}
	cfg.DocumentManifest = resolvePath(configPath, cfg.DocumentManifest)
	cfg.ChunkingConfig = resolvePath(configPath, cfg.ChunkingConfig)
	cfg.EmbeddingInput = resolvePath(configPath, cfg.EmbeddingInput)
	cfg.VectorIndex = resolvePath(configPath, cfg.VectorIndex)
	if cfg.CorpusSnapshotID == "" {
		return types.Statement{}, fmt.Errorf("corpus_snapshot_id is required")
	}
	for _, req := range []struct {
		path string
		name string
	}{{cfg.DocumentManifest, "document_manifest"}, {cfg.ChunkingConfig, "chunking_config"}, {cfg.EmbeddingInput, "embedding_input"}, {cfg.VectorIndex, "vector_index"}} {
		if err := requirePath(req.path, req.name); err != nil {
			return types.Statement{}, err
		}
	}
	if cfg.EmbeddingModel == "" {
		return types.Statement{}, fmt.Errorf("embedding_model is required")
	}
	if cfg.IndexBuilderImageDigest == "" {
		return types.Statement{}, fmt.Errorf("index_builder_image_digest is required")
	}

	connectorDigests := make([]types.NamedDigest, 0, len(cfg.ConnectorConfigs))
	subjects := make([]types.Subject, 0)
	for _, path := range cfg.ConnectorConfigs {
		if err := requirePath(path, "connector_config"); err != nil {
			return types.Statement{}, err
		}
		d, _, err := hash.DigestFile(path)
		if err != nil {
			return types.Statement{}, err
		}
		connectorDigests = append(connectorDigests, types.NamedDigest{Name: filepath.Base(path), Digest: d})
		s, err := subjectFromPath(path)
		if err != nil {
			return types.Statement{}, err
		}
		subjects = append(subjects, s)
	}

	docDigest, _, _ := hash.DigestFile(cfg.DocumentManifest)
	chunkDigest, _, _ := hash.DigestFile(cfg.ChunkingConfig)
	embedInputDigest, _, _ := hash.DigestFile(cfg.EmbeddingInput)
	vectorDigest, _, err := hash.DigestFile(cfg.VectorIndex)
	if err != nil {
		return types.Statement{}, err
	}

	predicate := types.CorpusPredicate{
		CorpusSnapshotID:        cfg.CorpusSnapshotID,
		ConnectorConfigDigests:  connectorDigests,
		DocumentManifestDigest:  docDigest,
		ChunkingConfigDigest:    chunkDigest,
		EmbeddingModel:          cfg.EmbeddingModel,
		EmbeddingInputDigest:    embedInputDigest,
		IndexBuilderImageDigest: cfg.IndexBuilderImageDigest,
		VectorIndexDigest:       vectorDigest,
	}
	if cfg.BuildCommand != "" {
		predicate.BuildCommandDigest = digestOfString(cfg.BuildCommand)
	}

	for _, p := range []string{cfg.DocumentManifest, cfg.ChunkingConfig, cfg.EmbeddingInput, cfg.VectorIndex} {
		s, err := subjectFromPath(p)
		if err != nil {
			return types.Statement{}, err
		}
		subjects = append(subjects, s)
	}
	return newStatement(types.AttestationCorpus, predicate, subjects, nil), nil
}
