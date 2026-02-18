package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/attest"
	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/hash"
	policyrego "github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/policy/rego"
	policyyaml "github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/policy/yaml"
	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/report"
	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/sign"
	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/store"
	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/verify"
	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/webhook"
	"github.com/spf13/cobra"
)

type cliError struct {
	code int
	err  error
}

func (e cliError) Error() string { return e.err.Error() }

func main() {
	root := newRootCommand()
	if err := root.Execute(); err != nil {
		var ce cliError
		if errors.As(err, &ce) {
			fmt.Fprintln(os.Stderr, ce.err)
			os.Exit(ce.code)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var ociPullFunc = store.PullOCI
var ociPublishFunc = store.PublishOCI

func newRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "llmsa",
		Short: "LLM Supply-Chain Attestation CLI",
	}
	root.AddCommand(newInitCommand())
	root.AddCommand(newAttestCommand())
	root.AddCommand(newSignCommand())
	root.AddCommand(newPublishCommand())
	root.AddCommand(newVerifyCommand())
	root.AddCommand(newGateCommand())
	root.AddCommand(newReportCommand())
	root.AddCommand(newDemoCommand())
	root.AddCommand(newWebhookCommand())
	return root
}

func newInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize llmsa configuration and local store",
		RunE: func(_ *cobra.Command, _ []string) error {
			if _, err := store.EnsureDefaultAttestationDir(); err != nil {
				return err
			}
			if !fileExists("llmsa.yaml") {
				if err := os.WriteFile("llmsa.yaml", []byte(defaultConfigYAML), 0o644); err != nil {
					return err
				}
			}
			if !fileExists("policy/examples/mvp-gates.yaml") {
				if err := os.MkdirAll("policy/examples", 0o755); err != nil {
					return err
				}
				if err := os.WriteFile("policy/examples/mvp-gates.yaml", []byte(defaultPolicyYAML), 0o644); err != nil {
					return err
				}
			}
			if !fileExists(".llmsa/dev_ed25519.pem") {
				if err := os.MkdirAll(".llmsa", 0o755); err != nil {
					return err
				}
				if err := sign.GeneratePEMPrivateKey(".llmsa/dev_ed25519.pem"); err != nil {
					return err
				}
			}
			fmt.Println("initialized llmsa config, policy, and local key")
			return nil
		},
	}
}

func newAttestCommand() *cobra.Command {
	attestCmd := &cobra.Command{Use: "attest", Short: "Create attestations"}

	var attType, cfgPath, outDir, gitRef string
	var changedOnly bool
	var determinismCheck int

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create statement attestation(s)",
		RunE: func(_ *cobra.Command, _ []string) error {
			if changedOnly {
				files, err := attest.CreateChangedOnly(gitRef, outDir, determinismCheck)
				if err != nil {
					return err
				}
				for _, f := range files {
					fmt.Println(f)
				}
				return nil
			}
			if attType == "" || cfgPath == "" {
				return fmt.Errorf("--type and --config are required when --changed-only is false")
			}
			files, err := attest.CreateByType(attest.CreateOptions{
				Type:             attType,
				ConfigPath:       cfgPath,
				OutDir:           outDir,
				DeterminismCheck: determinismCheck,
			})
			if err != nil {
				return err
			}
			for _, f := range files {
				fmt.Println(f)
			}
			return nil
		},
	}
	createCmd.Flags().StringVar(&attType, "type", "", "attestation type")
	createCmd.Flags().StringVar(&cfgPath, "config", "", "collector config file")
	createCmd.Flags().StringVar(&outDir, "out", ".llmsa/attestations", "output directory")
	createCmd.Flags().BoolVar(&changedOnly, "changed-only", false, "create attestations from changed files")
	createCmd.Flags().StringVar(&gitRef, "git-ref", "HEAD~1", "git reference for changed-only")
	createCmd.Flags().IntVar(&determinismCheck, "determinism-check", 1, "run attest generation multiple times and compare hashes")

	attestCmd.AddCommand(createCmd)
	return attestCmd
}

func newSignCommand() *cobra.Command {
	var inPath, provider, outPath, keyPath, oidcIssuer, oidcIdentity string
	cmd := &cobra.Command{
		Use:   "sign",
		Short: "Sign a statement and emit DSSE bundle",
		RunE: func(_ *cobra.Command, _ []string) error {
			if inPath == "" {
				return fmt.Errorf("--in is required")
			}
			raw, err := os.ReadFile(inPath)
			if err != nil {
				return err
			}
			var statement map[string]any
			if err := json.Unmarshal(raw, &statement); err != nil {
				return err
			}

			canonical, err := canonicalPayload(statement)
			if err != nil {
				return err
			}

			var material sign.SignMaterial
			switch provider {
			case "pem":
				if keyPath == "" {
					return fmt.Errorf("--key is required for pem provider")
				}
				signer, err := sign.NewPEMSigner(keyPath)
				if err != nil {
					return err
				}
				material, err = signer.Sign(canonical)
				if err != nil {
					return err
				}
			case "sigstore":
				signer := &sign.SigstoreSigner{PEMKeyPath: keyPath, Issuer: oidcIssuer, Identity: oidcIdentity}
				material, err = signer.Sign(canonical)
				if err != nil {
					return err
				}
			case "kms":
				signer := &sign.KMSSigner{}
				_, err := signer.Sign(canonical)
				return err
			default:
				return fmt.Errorf("unsupported provider %s", provider)
			}

			bundle, err := sign.CreateBundle(statement, material)
			if err != nil {
				return err
			}
			if outPath == "" {
				outPath = defaultBundlePath(inPath, statement)
			} else if fi, err := os.Stat(outPath); err == nil && fi.IsDir() {
				outPath = filepath.Join(outPath, filepath.Base(defaultBundlePath(inPath, statement)))
			}
			if err := sign.WriteBundle(outPath, bundle); err != nil {
				return err
			}
			fmt.Println(outPath)
			return nil
		},
	}
	cmd.Flags().StringVar(&inPath, "in", "", "statement JSON input")
	cmd.Flags().StringVar(&provider, "provider", "sigstore", "signing provider (sigstore|pem|kms)")
	cmd.Flags().StringVar(&outPath, "out", "", "bundle output path")
	cmd.Flags().StringVar(&keyPath, "key", "", "PEM key path")
	cmd.Flags().StringVar(&oidcIssuer, "oidc-issuer", "", "sigstore OIDC issuer")
	cmd.Flags().StringVar(&oidcIdentity, "oidc-identity", "", "sigstore OIDC identity")
	return cmd
}

func newPublishCommand() *cobra.Command {
	var inPath, ociRef string
	cmd := &cobra.Command{
		Use:   "publish",
		Short: "Publish DSSE bundle to OCI",
		RunE: func(_ *cobra.Command, _ []string) error {
			if inPath == "" || ociRef == "" {
				return fmt.Errorf("--in and --oci are required")
			}
			pinned, err := ociPublishFunc(inPath, ociRef)
			if err != nil {
				return err
			}
			fmt.Println(pinned)
			return nil
		},
	}
	cmd.Flags().StringVar(&inPath, "in", "", "bundle path")
	cmd.Flags().StringVar(&ociRef, "oci", "", "OCI destination")
	return cmd
}

func newVerifyCommand() *cobra.Command {
	var sourceType, sourcePath, policyPath, format, outPath, schemaDir string
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify bundle signatures, schemas, and digests",
		RunE: func(_ *cobra.Command, _ []string) error {
			if sourcePath == "" {
				sourcePath = ".llmsa/attestations"
			}
			if schemaDir == "" {
				schemaDir = "schemas/v1"
			}
			signerPolicy := verify.SignerPolicy{}
			if policyPath != "" {
				pol, err := policyyaml.LoadPolicy(policyPath)
				if err != nil {
					return err
				}
				signerPolicy.OIDCIssuer = pol.OIDCIssuer
				signerPolicy.IdentityRegex = pol.IdentityRegex
			}

			resolvedSource := sourcePath
			if sourceType == "oci" {
				tmpDir, err := os.MkdirTemp("", "llmsa-oci-verify-")
				if err != nil {
					return err
				}
				defer os.RemoveAll(tmpDir)
				refs := splitCSV(sourcePath)
				if len(refs) == 0 {
					return fmt.Errorf("--attestations must include at least one OCI ref for --source oci")
				}
				for i, ref := range refs {
					out := filepath.Join(tmpDir, fmt.Sprintf("oci_%d.bundle.json", i+1))
					if err := ociPullFunc(ref, out); err != nil {
						return err
					}
				}
				resolvedSource = tmpDir
			} else if sourceType != "local" {
				return fmt.Errorf("unsupported source %s", sourceType)
			}

			r := verify.Run(verify.Options{SourcePath: resolvedSource, SchemaDir: schemaDir, SignerPolicy: signerPolicy})

			switch format {
			case "json":
				if outPath == "" {
					outPath = "verify.json"
				}
				if err := report.WriteJSON(outPath, r); err != nil {
					return err
				}
				fmt.Println(outPath)
			case "md":
				if outPath == "" {
					outPath = "verify.md"
				}
				if err := report.WriteMarkdown(outPath, r); err != nil {
					return err
				}
				fmt.Println(outPath)
			default:
				return fmt.Errorf("unsupported format %s", format)
			}

			if !r.Passed {
				return cliError{code: r.ExitCode, err: fmt.Errorf("verification failed")}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&sourceType, "source", "local", "source type (local|oci)")
	cmd.Flags().StringVar(&sourcePath, "attestations", ".llmsa/attestations", "bundle path or directory")
	cmd.Flags().StringVar(&policyPath, "policy", "", "policy yaml path")
	cmd.Flags().StringVar(&format, "format", "json", "output format (json|md)")
	cmd.Flags().StringVar(&outPath, "out", "", "output report path")
	cmd.Flags().StringVar(&schemaDir, "schema-dir", "schemas/v1", "schema directory")
	return cmd
}

func newGateCommand() *cobra.Command {
	var policyPath, attestationsPath, gitRef, sourceType, engine, regoPolicyPath string
	cmd := &cobra.Command{
		Use:   "gate",
		Short: "Run policy gates and return non-zero on violations",
		RunE: func(_ *cobra.Command, _ []string) error {
			if policyPath == "" {
				return fmt.Errorf("--policy is required")
			}
			if attestationsPath == "" {
				attestationsPath = ".llmsa/attestations"
			}
			resolvedSource := attestationsPath
			if sourceType == "oci" {
				tmpDir, err := os.MkdirTemp("", "llmsa-oci-gate-")
				if err != nil {
					return err
				}
				defer os.RemoveAll(tmpDir)
				refs := splitCSV(attestationsPath)
				if len(refs) == 0 {
					return fmt.Errorf("--attestations must include at least one OCI ref for --source oci")
				}
				for i, ref := range refs {
					out := filepath.Join(tmpDir, fmt.Sprintf("oci_%d.bundle.json", i+1))
					if err := ociPullFunc(ref, out); err != nil {
						return err
					}
				}
				resolvedSource = tmpDir
			} else if sourceType != "local" {
				return fmt.Errorf("unsupported source %s", sourceType)
			}
			policy, err := policyyaml.LoadPolicy(policyPath)
			if err != nil {
				return err
			}
			statements, err := policyyaml.LoadStatements(resolvedSource)
			if err != nil {
				return err
			}
			violations := []string{}
			switch engine {
			case "yaml":
				violations, err = policyyaml.Evaluate(policy, statements, gitRef)
				if err != nil {
					return err
				}
			case "rego":
				changed, err := policyyaml.ChangedFiles(gitRef)
				if err != nil {
					return err
				}
				result, err := policyrego.Evaluate(regoPolicyPath, policyrego.BuildInput(policy, statements, changed))
				if err != nil {
					return err
				}
				if !result.Allow {
					violations = append(violations, result.Violations...)
					if len(violations) == 0 {
						violations = append(violations, "rego policy denied request")
					}
				}
			default:
				return fmt.Errorf("unsupported policy engine %s", engine)
			}
			if len(violations) > 0 {
				for _, v := range violations {
					fmt.Println(v)
				}
				return cliError{code: verify.ExitPolicyFail, err: fmt.Errorf("policy gate failed")}
			}
			fmt.Println("policy gate passed")
			return nil
		},
	}
	cmd.Flags().StringVar(&policyPath, "policy", "", "policy YAML path")
	cmd.Flags().StringVar(&attestationsPath, "attestations", ".llmsa/attestations", "attestation directory or file")
	cmd.Flags().StringVar(&sourceType, "source", "local", "attestation source type (local|oci)")
	cmd.Flags().StringVar(&engine, "engine", "yaml", "policy engine (yaml|rego)")
	cmd.Flags().StringVar(&regoPolicyPath, "rego-policy", "policy/examples/rego-gates.rego", "rego policy path (used with --engine rego)")
	cmd.Flags().StringVar(&gitRef, "git-ref", "HEAD~1", "git reference for changed-file triggers")
	return cmd
}

func newReportCommand() *cobra.Command {
	var inPath, outPath string
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate markdown report from verify JSON",
		RunE: func(_ *cobra.Command, _ []string) error {
			if inPath == "" || outPath == "" {
				return fmt.Errorf("--in and --out are required")
			}
			raw, err := os.ReadFile(inPath)
			if err != nil {
				return err
			}
			var r verify.Report
			if err := json.Unmarshal(raw, &r); err != nil {
				return err
			}
			if err := report.WriteMarkdown(outPath, r); err != nil {
				return err
			}
			fmt.Println(outPath)
			return nil
		},
	}
	cmd.Flags().StringVar(&inPath, "in", "", "verify report json input")
	cmd.Flags().StringVar(&outPath, "out", "", "markdown output")
	return cmd
}

func newDemoCommand() *cobra.Command {
	demoCmd := &cobra.Command{Use: "demo", Short: "Demo operations"}
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run tiny-rag demo sequence",
		RunE: func(_ *cobra.Command, _ []string) error {
			cmd := exec.Command("make", "-C", "examples/tiny-rag", "demo")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		},
	}
	demoCmd.AddCommand(runCmd)
	return demoCmd
}

func canonicalPayload(statement map[string]any) ([]byte, error) {
	return hash.CanonicalJSON(statement)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func defaultBundlePath(inPath string, statement map[string]any) string {
	attType := asString(statement["attestation_type"])
	statementID := asString(statement["statement_id"])
	gitSHA := "local"
	if g, ok := statement["generator"].(map[string]any); ok {
		if v, ok := g["git_sha"].(string); ok && v != "" {
			gitSHA = v
		}
	}
	gitSHA = strings.NewReplacer("/", "_", ":", "_", " ", "_").Replace(gitSHA)
	name := fmt.Sprintf("attestation_%s_%s_%s.bundle.json", attType, gitSHA, statementID)
	return filepath.Join(filepath.Dir(inPath), name)
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

const defaultConfigYAML = `collectors:
  prompt_attestation: examples/tiny-rag/configs/prompt.yaml
  corpus_attestation: examples/tiny-rag/configs/corpus.yaml
  eval_attestation: examples/tiny-rag/configs/eval.yaml
  route_attestation: examples/tiny-rag/configs/route.yaml
  slo_attestation: examples/tiny-rag/configs/slo.yaml
path_rules:
  prompt_attestation:
    - examples/tiny-rag/app/**
  corpus_attestation:
    - examples/tiny-rag/data/**
  eval_attestation:
    - examples/tiny-rag/eval/**
  route_attestation:
    - examples/tiny-rag/route/**
  slo_attestation:
    - examples/tiny-rag/slo/**
`

const defaultPolicyYAML = `version: 1
oidc_issuer: https://token.actions.githubusercontent.com
identity_regex: '^https://github\.com/.+/.+/.github/workflows/.+@refs/.+$'
plaintext_allowlist: []
gates:
  - id: G001
    trigger_paths: ["examples/tiny-rag/app/**"]
    required_attestations: ["prompt_attestation", "eval_attestation"]
    message: "Prompt changed without passing eval attestation."
  - id: G002
    trigger_paths: ["examples/tiny-rag/data/**"]
    required_attestations: ["corpus_attestation", "eval_attestation"]
    message: "Corpus changed without rebuild+eval attestations."
  - id: G003
    trigger_paths: ["examples/tiny-rag/route/**"]
    required_attestations: ["route_attestation", "slo_attestation"]
    message: "Route changed without valid SLO attestation."
  - id: G004
    trigger_paths: ["examples/tiny-rag/eval/**"]
    required_attestations: ["eval_attestation"]
    message: "Eval config changed without signed eval attestation."
  - id: G005
    trigger_paths: ["refs/tags/v*"]
    required_attestations: ["prompt_attestation", "corpus_attestation", "eval_attestation", "route_attestation", "slo_attestation"]
    message: "Release blocked: incomplete attestation set."
  - id: G006
    trigger_paths: ["**"]
    required_attestations: []
    message: "Sensitive payload exposure blocked by policy."
`

func newWebhookCommand() *cobra.Command {
	webhookCmd := &cobra.Command{
		Use:   "webhook",
		Short: "Kubernetes admission webhook",
	}

	var port int
	var tlsCert, tlsKey, policy, schemaDir, registryPrefix string
	var failOpen bool
	var cacheTTLSeconds int

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the validating admission webhook server",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg := webhook.Config{
				Port:            port,
				TLSCertPath:     tlsCert,
				TLSKeyPath:      tlsKey,
				PolicyPath:      policy,
				SchemaDir:       schemaDir,
				RegistryPrefix:  registryPrefix,
				FailOpen:        failOpen,
				CacheTTLSeconds: cacheTTLSeconds,
			}
			mux := http.NewServeMux()
			mux.Handle("/validate", webhook.Handler(cfg))
			mux.Handle("/healthz", webhook.HealthHandler())

			addr := fmt.Sprintf(":%d", cfg.Port)
			fmt.Fprintf(os.Stderr, "webhook listening on %s\n", addr)
			if cfg.TLSCertPath != "" && cfg.TLSKeyPath != "" {
				return http.ListenAndServeTLS(addr, cfg.TLSCertPath, cfg.TLSKeyPath, mux)
			}
			return http.ListenAndServe(addr, mux)
		},
	}
	serveCmd.Flags().IntVar(&port, "port", 8443, "webhook listen port")
	serveCmd.Flags().StringVar(&tlsCert, "tls-cert", "", "TLS certificate path")
	serveCmd.Flags().StringVar(&tlsKey, "tls-key", "", "TLS key path")
	serveCmd.Flags().StringVar(&policy, "policy", "", "policy YAML path")
	serveCmd.Flags().StringVar(&schemaDir, "schema-dir", "schemas/v1", "schema directory")
	serveCmd.Flags().StringVar(&registryPrefix, "registry-prefix", "", "OCI registry prefix for attestation bundles")
	serveCmd.Flags().BoolVar(&failOpen, "fail-open", false, "allow pods when verification encounters an error")
	serveCmd.Flags().IntVar(&cacheTTLSeconds, "cache-ttl-seconds", 300, "successful verification cache TTL in seconds")

	webhookCmd.AddCommand(serveCmd)
	return webhookCmd
}
