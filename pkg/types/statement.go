package types

type Statement struct {
	SchemaVersion   string            `json:"schema_version"`
	StatementID     string            `json:"statement_id"`
	AttestationType string            `json:"attestation_type"`
	PredicateType   string            `json:"predicate_type"`
	GeneratedAt     string            `json:"generated_at"`
	Generator       Generator         `json:"generator"`
	Subject         []Subject         `json:"subject"`
	Materials       []Subject         `json:"materials,omitempty"`
	Predicate       any               `json:"predicate"`
	Privacy         Privacy           `json:"privacy"`
	Annotations     map[string]string `json:"annotations,omitempty"`
}

type Generator struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	GitSHA  string `json:"git_sha"`
}

type Subject struct {
	Name      string `json:"name"`
	URI       string `json:"uri"`
	Digest    Digest `json:"digest"`
	SizeBytes int64  `json:"size_bytes"`
}

type Digest struct {
	SHA256 string `json:"sha256"`
}

type Privacy struct {
	Mode                           string `json:"mode"`
	EncryptedBlobDigest            string `json:"encrypted_blob_digest,omitempty"`
	EncryptionRecipientFingerprint string `json:"encryption_recipient_fingerprint,omitempty"`
}

const (
	AttestationPrompt = "prompt_attestation"
	AttestationCorpus = "corpus_attestation"
	AttestationEval   = "eval_attestation"
	AttestationRoute  = "route_attestation"
	AttestationSLO    = "slo_attestation"
)

func PredicateURI(attestationType string) string {
	switch attestationType {
	case AttestationPrompt:
		return "https://llmsa.dev/attestation/prompt/v1"
	case AttestationCorpus:
		return "https://llmsa.dev/attestation/corpus/v1"
	case AttestationEval:
		return "https://llmsa.dev/attestation/eval/v1"
	case AttestationRoute:
		return "https://llmsa.dev/attestation/route/v1"
	case AttestationSLO:
		return "https://llmsa.dev/attestation/slo/v1"
	default:
		return ""
	}
}
