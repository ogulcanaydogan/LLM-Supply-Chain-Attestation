package attest

import (
	"fmt"

	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/hash"
	"github.com/ogulcanaydogan/llm-supply-chain-attestation/pkg/types"
)

type PromptConfig struct {
	SystemPrompt      string   `yaml:"system_prompt"`
	TemplatesDir      string   `yaml:"templates_dir"`
	ToolSchemasDir    string   `yaml:"tool_schemas_dir"`
	SafetyPolicy      string   `yaml:"safety_policy"`
	RenderConfig      string   `yaml:"render_config"`
	TestSuite         string   `yaml:"test_suite"`
	SensitivityLabels []string `yaml:"sensitivity_labels"`
}

func CollectPrompt(configPath string) (types.Statement, error) {
	cfg := PromptConfig{}
	if err := LoadConfig(configPath, &cfg); err != nil {
		return types.Statement{}, err
	}
	cfg.SystemPrompt = resolvePath(configPath, cfg.SystemPrompt)
	cfg.TemplatesDir = resolvePath(configPath, cfg.TemplatesDir)
	cfg.ToolSchemasDir = resolvePath(configPath, cfg.ToolSchemasDir)
	cfg.SafetyPolicy = resolvePath(configPath, cfg.SafetyPolicy)
	cfg.RenderConfig = resolvePath(configPath, cfg.RenderConfig)
	cfg.TestSuite = resolvePath(configPath, cfg.TestSuite)
	for _, req := range []struct {
		path string
		name string
	}{{cfg.SystemPrompt, "system_prompt"}, {cfg.TemplatesDir, "templates_dir"}, {cfg.ToolSchemasDir, "tool_schemas_dir"}, {cfg.SafetyPolicy, "safety_policy"}} {
		if err := requirePath(req.path, req.name); err != nil {
			return types.Statement{}, err
		}
	}

	systemDigest, _, err := hash.DigestFile(cfg.SystemPrompt)
	if err != nil {
		return types.Statement{}, fmt.Errorf("digest system prompt: %w", err)
	}
	templateDigests, templateSubjects, err := sortedFileDigests(cfg.TemplatesDir)
	if err != nil {
		return types.Statement{}, fmt.Errorf("digest templates: %w", err)
	}
	toolDigests, toolSubjects, err := sortedFileDigests(cfg.ToolSchemasDir)
	if err != nil {
		return types.Statement{}, fmt.Errorf("digest tool schemas: %w", err)
	}
	safetyDigest, _, err := hash.DigestFile(cfg.SafetyPolicy)
	if err != nil {
		return types.Statement{}, fmt.Errorf("digest safety policy: %w", err)
	}

	predicate := types.PromptPredicate{
		PromptBundleDigest: bundleDigest(systemDigest, safetyDigest, bundleDigest(templateDigests...), bundleDigest(toolDigests...)),
		SystemPromptDigest: systemDigest,
		TemplateDigests:    templateDigests,
		ToolSchemaDigests:  toolDigests,
		SafetyPolicyDigest: safetyDigest,
		SensitivityLabels:  cfg.SensitivityLabels,
	}
	if cfg.RenderConfig != "" {
		d, _, err := hash.DigestFile(cfg.RenderConfig)
		if err != nil {
			return types.Statement{}, err
		}
		predicate.PromptRenderConfigDigest = d
	}
	if cfg.TestSuite != "" {
		d, _, err := hash.DigestFile(cfg.TestSuite)
		if err != nil {
			return types.Statement{}, err
		}
		predicate.PromptTestSuiteDigest = d
	}

	subjects := make([]types.Subject, 0, 2+len(templateSubjects)+len(toolSubjects))
	sysSubject, err := subjectFromPath(cfg.SystemPrompt)
	if err != nil {
		return types.Statement{}, err
	}
	safetySubject, err := subjectFromPath(cfg.SafetyPolicy)
	if err != nil {
		return types.Statement{}, err
	}
	subjects = append(subjects, sysSubject, safetySubject)
	subjects = append(subjects, templateSubjects...)
	subjects = append(subjects, toolSubjects...)

	return newStatement(types.AttestationPrompt, predicate, subjects, nil), nil
}
