package types

type PromptPredicate struct {
	PromptBundleDigest       string   `json:"prompt_bundle_digest"`
	SystemPromptDigest       string   `json:"system_prompt_digest"`
	TemplateDigests          []string `json:"template_digests"`
	ToolSchemaDigests        []string `json:"tool_schema_digests"`
	SafetyPolicyDigest       string   `json:"safety_policy_digest"`
	PromptRenderConfigDigest string   `json:"prompt_render_config_digest,omitempty"`
	PromptTestSuiteDigest    string   `json:"prompt_test_suite_digest,omitempty"`
	SensitivityLabels        []string `json:"sensitivity_labels,omitempty"`
}
