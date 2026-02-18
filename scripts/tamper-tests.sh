#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

BIN=".llmsa/llmsa"
ATTEST_DIR=".llmsa/attestations"
OUT_DIR=".llmsa/tamper"
RESULTS_CSV="${OUT_DIR}/results.csv"
RESULTS_JSON="${OUT_DIR}/results.json"
RESULTS_MD="${OUT_DIR}/results.md"

mkdir -p "${OUT_DIR}"
echo "id,name,expected_exit,actual_exit,status" > "${RESULTS_CSV}"

go build -o "${BIN}" ./cmd/llmsa
make -C examples/tiny-rag clean attest sign >/dev/null

bundle_for_type() {
  local typ="$1"
  ls "${ATTEST_DIR}/attestation_${typ}"_*.bundle.json | head -n1
}

PROMPT_BUNDLE="$(bundle_for_type prompt_attestation)"
CORPUS_BUNDLE="$(bundle_for_type corpus_attestation)"
EVAL_BUNDLE="$(bundle_for_type eval_attestation)"
ROUTE_BUNDLE="$(bundle_for_type route_attestation)"
SLO_BUNDLE="$(bundle_for_type slo_attestation)"

run_verify_code() {
  local source_path="$1"
  local out_file="$2"
  set +e
  "${BIN}" verify \
    --source local \
    --attestations "${source_path}" \
    --policy policy/examples/mvp-gates.yaml \
    --format json \
    --out "${out_file}" >/dev/null 2>&1
  local code=$?
  set -e
  echo "${code}"
}

record_result() {
  local id="$1"
  local name="$2"
  local expected="$3"
  local actual="$4"
  local status="FAIL"
  if [[ "${expected}" == "${actual}" ]]; then
    status="PASS"
  fi
  echo "${id},${name},${expected},${actual},${status}" >> "${RESULTS_CSV}"
  if [[ "${status}" == "PASS" ]]; then
    echo "[PASS] ${id} ${name} (exit=${actual})"
  else
    echo "[FAIL] ${id} ${name} expected=${expected} actual=${actual}"
  fi
}

decode_bundle_statement() {
  local bundle_path="$1"
  local statement_path="$2"
  python3 - "${bundle_path}" "${statement_path}" <<'PY'
import base64
import json
import sys

bundle_path, out_path = sys.argv[1], sys.argv[2]
with open(bundle_path, "r", encoding="utf-8") as f:
    bundle = json.load(f)
payload = base64.b64decode(bundle["envelope"]["payload"])
statement = json.loads(payload.decode("utf-8"))
with open(out_path, "w", encoding="utf-8") as f:
    json.dump(statement, f, indent=2, sort_keys=True)
PY
}

mutate_statement_json() {
  local statement_path="$1"
  local mutation="$2"
  python3 - "${statement_path}" "${mutation}" <<'PY'
import json
import sys

statement_path, mutation = sys.argv[1], sys.argv[2]
with open(statement_path, "r", encoding="utf-8") as f:
    statement = json.load(f)

if mutation == "remove_prompt_required_field":
    statement["predicate"].pop("system_prompt_digest", None)
elif mutation == "invalid_generated_at":
    statement["generated_at"] = "invalid-time"
elif mutation == "invalid_subject_digest":
    statement["subject"][0]["digest"]["sha256"] = "abc"
elif mutation == "unknown_depends_on":
    ann = statement.setdefault("annotations", {})
    ann["depends_on"] = "unknown-ref"
else:
    raise SystemExit(f"unknown mutation: {mutation}")

with open(statement_path, "w", encoding="utf-8") as f:
    json.dump(statement, f, indent=2, sort_keys=True)
PY
}

resign_statement() {
  local statement_path="$1"
  local output_bundle="$2"
  "${BIN}" sign \
    --in "${statement_path}" \
    --provider sigstore \
    --key .llmsa/dev_ed25519.pem \
    --oidc-issuer https://token.actions.githubusercontent.com \
    --oidc-identity https://github.com/local/dev/.github/workflows/manual.yml@refs/heads/local \
    --out "${output_bundle}" >/dev/null
}

tamper_bundle_json() {
  local src_bundle="$1"
  local dst_bundle="$2"
  local mutation="$3"
  python3 - "${src_bundle}" "${dst_bundle}" "${mutation}" <<'PY'
import json
import sys

src, dst, mutation = sys.argv[1], sys.argv[2], sys.argv[3]
with open(src, "r", encoding="utf-8") as f:
    bundle = json.load(f)

if mutation == "sig":
    bundle["envelope"]["signatures"][0]["sig"] = "AAAA" + bundle["envelope"]["signatures"][0]["sig"][4:]
elif mutation == "pubkey":
    bundle["envelope"]["signatures"][0]["public_key_pem"] = "tampered-key"
elif mutation == "statement_hash":
    bundle["metadata"]["statement_hash"] = "sha256:" + "0" * 64
elif mutation == "remove_signatures":
    bundle["envelope"]["signatures"] = []
else:
    raise SystemExit(f"unknown mutation: {mutation}")

with open(dst, "w", encoding="utf-8") as f:
    json.dump(bundle, f, indent=2)
PY
}

run_file_mutation_case() {
  local id="$1"
  local name="$2"
  local path="$3"
  local backup="${OUT_DIR}/${id}.bak"
  cp "${path}" "${backup}"
  echo "# tamper-${id}" >> "${path}"
  local code
  code="$(run_verify_code "${ATTEST_DIR}" "${OUT_DIR}/verify-${id}.json")"
  mv "${backup}" "${path}"
  record_result "${id}" "${name}" "12" "${code}"
}

run_bundle_mutation_case() {
  local id="$1"
  local name="$2"
  local mutation="$3"
  local case_dir="${OUT_DIR}/case-${id}"
  mkdir -p "${case_dir}"
  local tampered="${case_dir}/prompt.bundle.json"
  tamper_bundle_json "${PROMPT_BUNDLE}" "${tampered}" "${mutation}"
  local code
  code="$(run_verify_code "${case_dir}" "${OUT_DIR}/verify-${id}.json")"
  record_result "${id}" "${name}" "11" "${code}"
}

run_schema_mutation_case() {
  local id="$1"
  local name="$2"
  local mutation="$3"
  local case_dir="${OUT_DIR}/case-${id}"
  mkdir -p "${case_dir}"
  local statement="${case_dir}/statement.json"
  local bundle="${case_dir}/mutated.bundle.json"
  decode_bundle_statement "${PROMPT_BUNDLE}" "${statement}"
  mutate_statement_json "${statement}" "${mutation}"
  resign_statement "${statement}" "${bundle}"
  local code
  code="$(run_verify_code "${case_dir}" "${OUT_DIR}/verify-${id}.json")"
  record_result "${id}" "${name}" "14" "${code}"
}

run_chain_subset_case() {
  local id="$1"
  local name="$2"
  shift 2
  local case_dir="${OUT_DIR}/case-${id}"
  mkdir -p "${case_dir}"
  for src in "$@"; do
    cp "${src}" "${case_dir}/"
  done
  local code
  code="$(run_verify_code "${case_dir}" "${OUT_DIR}/verify-${id}.json")"
  record_result "${id}" "${name}" "14" "${code}"
}

run_chain_mutation_case() {
  local id="$1"
  local name="$2"
  local case_dir="${OUT_DIR}/case-${id}"
  mkdir -p "${case_dir}"
  cp "${PROMPT_BUNDLE}" "${case_dir}/"
  cp "${CORPUS_BUNDLE}" "${case_dir}/"
  cp "${EVAL_BUNDLE}" "${case_dir}/"
  local route_copy="${case_dir}/route.bundle.json"
  cp "${ROUTE_BUNDLE}" "${route_copy}"
  local statement="${case_dir}/route.statement.json"
  decode_bundle_statement "${route_copy}" "${statement}"
  mutate_statement_json "${statement}" "unknown_depends_on"
  resign_statement "${statement}" "${route_copy}"
  local code
  code="$(run_verify_code "${case_dir}" "${OUT_DIR}/verify-${id}.json")"
  record_result "${id}" "${name}" "14" "${code}"
}

# T01-T10: subject/material byte mutations (digest mismatch -> exit 12)
run_file_mutation_case "T01" "prompt_system_prompt_byte_flip" "examples/tiny-rag/app/system_prompt.txt"
run_file_mutation_case "T02" "prompt_template_byte_flip" "examples/tiny-rag/app/templates/base.prompt.txt"
run_file_mutation_case "T03" "prompt_tool_schema_byte_flip" "examples/tiny-rag/app/tools/search.schema.json"
run_file_mutation_case "T04" "prompt_safety_policy_byte_flip" "examples/tiny-rag/app/safety-policy.yaml"
run_file_mutation_case "T05" "corpus_document_manifest_byte_flip" "examples/tiny-rag/data/document-manifest.json"
run_file_mutation_case "T06" "corpus_chunking_config_byte_flip" "examples/tiny-rag/data/chunking.yaml"
run_file_mutation_case "T07" "corpus_embedding_input_byte_flip" "examples/tiny-rag/data/embedding-input.jsonl"
run_file_mutation_case "T08" "eval_testset_byte_flip" "examples/tiny-rag/eval/testset.json"
run_file_mutation_case "T09" "route_config_byte_flip" "examples/tiny-rag/route/route.yaml"
run_file_mutation_case "T10" "slo_query_profile_byte_flip" "examples/tiny-rag/slo/profile.json"

# T11-T14: signature/bundle tampering (signature failure -> exit 11)
run_bundle_mutation_case "T11" "bundle_signature_corruption" "sig"
run_bundle_mutation_case "T12" "bundle_public_key_corruption" "pubkey"
run_bundle_mutation_case "T13" "bundle_statement_hash_corruption" "statement_hash"
run_bundle_mutation_case "T14" "bundle_signatures_removed" "remove_signatures"

# T15-T17: schema/statement tampering with valid re-sign (schema failure -> exit 14)
run_schema_mutation_case "T15" "schema_missing_required_predicate_field" "remove_prompt_required_field"
run_schema_mutation_case "T16" "schema_invalid_generated_at" "invalid_generated_at"
run_schema_mutation_case "T17" "schema_invalid_subject_digest_format" "invalid_subject_digest"

# T18-T20: provenance chain integrity tampering (chain failure -> exit 14)
run_chain_subset_case "T18" "chain_route_slo_without_eval" "${ROUTE_BUNDLE}" "${SLO_BUNDLE}"
run_chain_subset_case "T19" "chain_eval_without_prompt_corpus" "${EVAL_BUNDLE}"
run_chain_mutation_case "T20" "chain_unknown_dependency_reference"

python3 - "${RESULTS_CSV}" "${RESULTS_JSON}" "${RESULTS_MD}" <<'PY'
import csv
import json
import sys
from pathlib import Path

csv_path = Path(sys.argv[1])
json_path = Path(sys.argv[2])
md_path = Path(sys.argv[3])

rows = list(csv.DictReader(csv_path.open()))
passed = sum(1 for row in rows if row["status"] == "PASS")
failed = len(rows) - passed

json_path.write_text(
    json.dumps(
        {
            "total_cases": len(rows),
            "passed": passed,
            "failed": failed,
            "results": rows,
        },
        indent=2,
    )
)

lines = [
    "# Tamper Suite Results",
    "",
    f"- Total: `{len(rows)}`",
    f"- Passed: `{passed}`",
    f"- Failed: `{failed}`",
    "",
    "| ID | Name | Expected Exit | Actual Exit | Status |",
    "|---|---|---:|---:|---|",
]
for row in rows:
    lines.append(
        f"| {row['id']} | {row['name']} | {row['expected_exit']} | {row['actual_exit']} | {row['status']} |"
    )
md_path.write_text("\n".join(lines) + "\n")
PY

FAILED_COUNT="$(awk -F, 'NR>1 && $5=="FAIL" {c++} END {print c+0}' "${RESULTS_CSV}")"
if [[ "${FAILED_COUNT}" -ne 0 ]]; then
  echo "tamper suite failed with ${FAILED_COUNT} failing cases"
  exit 1
fi

echo "tamper suite passed (20/20)"
