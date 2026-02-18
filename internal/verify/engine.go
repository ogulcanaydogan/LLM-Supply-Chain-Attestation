package verify

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/sign"
)

type Options struct {
	SourcePath   string
	SchemaDir    string
	SignerPolicy SignerPolicy
}

func Run(opts Options) Report {
	report := Report{Passed: true, ExitCode: ExitPass}
	paths, err := bundlePaths(opts.SourcePath)
	if err != nil {
		report.Passed = false
		report.ExitCode = ExitMissing
		report.Violations = append(report.Violations, err.Error())
		return report
	}
	if len(paths) == 0 {
		report.Passed = false
		report.ExitCode = ExitMissing
		report.Violations = append(report.Violations, "no bundle files found")
		return report
	}
	report.BundleCount = len(paths)

	for _, p := range paths {
		bundle, err := sign.ReadBundle(p)
		if err != nil {
			report.addFailure(p, "bundle_read", ExitMissing, err)
			continue
		}
		if err := VerifySignature(bundle, opts.SignerPolicy); err != nil {
			report.addFailure(p, "signature", ExitSignatureFail, err)
			continue
		}
		report.Checks = append(report.Checks, CheckResult{Bundle: p, Check: "signature", Passed: true, Message: "ok"})

		var statement map[string]any
		if err := sign.DecodePayload(bundle, &statement); err != nil {
			report.addFailure(p, "payload_decode", ExitMissing, err)
			continue
		}

		if err := VerifySchemas(opts.SchemaDir, statement); err != nil {
			report.addFailure(p, "schema", ExitSchemaFail, err)
			continue
		}
		report.Checks = append(report.Checks, CheckResult{Bundle: p, Check: "schema", Passed: true, Message: "ok"})

		if err := VerifyBasicChainConstraints(statement); err != nil {
			report.addFailure(p, "chain", ExitSchemaFail, err)
			continue
		}
		report.Checks = append(report.Checks, CheckResult{Bundle: p, Check: "chain", Passed: true, Message: "ok"})

		if err := VerifySubjects(statement); err != nil {
			report.addFailure(p, "subject_digest", ExitDigestMismatch, err)
			continue
		}
		report.Checks = append(report.Checks, CheckResult{Bundle: p, Check: "subject_digest", Passed: true, Message: "ok"})

		report.Statements = append(report.Statements, StatementSummary{
			AttestationType: asString(statement["attestation_type"]),
			StatementID:     asString(statement["statement_id"]),
			PrivacyMode:     privacyMode(statement),
		})
	}

	if report.Passed {
		report.ExitCode = ExitPass
	}
	return report
}

func (r *Report) addFailure(bundle, check string, exit int, err error) {
	r.Passed = false
	if r.ExitCode == ExitPass || exit > r.ExitCode {
		r.ExitCode = exit
	}
	msg := err.Error()
	r.Checks = append(r.Checks, CheckResult{Bundle: bundle, Check: check, Passed: false, Message: msg})
	r.Violations = append(r.Violations, fmt.Sprintf("%s: %s", check, msg))
}

func WriteJSON(path string, report Report) error {
	raw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

func bundlePaths(source string) ([]string, error) {
	fi, err := os.Stat(source)
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return []string{source}, nil
	}
	files := make([]string, 0)
	entries, err := os.ReadDir(source)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".bundle.json") {
			files = append(files, filepath.Join(source, name))
		}
	}
	sort.Strings(files)
	return files, nil
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}

func privacyMode(statement map[string]any) string {
	p, _ := statement["privacy"].(map[string]any)
	return asString(p["mode"])
}
