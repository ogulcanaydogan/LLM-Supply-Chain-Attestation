#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

BIN=".llmsa/llmsa"
TIMESTAMP="$(date -u +"%Y%m%dT%H%M%SZ")"
RUN_DIR=".llmsa/benchmarks/${TIMESTAMP}"
RAW_DIR="${RUN_DIR}/raw"
CSV_FILE="${RAW_DIR}/timings.csv"
SUMMARY_MD="${RUN_DIR}/summary.md"
MANIFEST_JSON="${RAW_DIR}/run-manifest.json"
REPRO_JSON="${RAW_DIR}/reproducibility.json"
DOCS_DIR="docs/benchmarks"
DOCS_FILE="${DOCS_DIR}/$(date -u +"%Y-%m").md"

mkdir -p "${RAW_DIR}" "${DOCS_DIR}"
echo "scenario,workload,iteration,duration_ms" > "${CSV_FILE}"

now_ns() {
  python3 - <<'PY'
import time
print(time.time_ns())
PY
}

record_timing() {
  local scenario="$1"
  local workload="$2"
  local iteration="$3"
  local duration_ms="$4"
  echo "${scenario},${workload},${iteration},${duration_ms}" >> "${CSV_FILE}"
}

measure_sign_and_verify() {
  local workload="$1"
  local statement_template="$2"
  local work_dir="${RUN_DIR}/w${workload}"
  local statement_dir="${work_dir}/statements"
  local bundle_dir="${work_dir}/bundles"
  mkdir -p "${statement_dir}" "${bundle_dir}"

  for i in $(seq 1 "${workload}"); do
    cp "${statement_template}" "${statement_dir}/s-${i}.json"
  done

  for iteration in 1 2 3; do
    rm -f "${bundle_dir}"/*.bundle.json || true

    local start_ns end_ns duration_ms
    start_ns="$(now_ns)"
    for stmt in "${statement_dir}"/*.json; do
      "${BIN}" sign \
        --in "${stmt}" \
        --provider sigstore \
        --key .llmsa/dev_ed25519.pem \
        --oidc-issuer https://token.actions.githubusercontent.com \
        --oidc-identity https://github.com/local/dev/.github/workflows/manual.yml@refs/heads/local \
        --out "${bundle_dir}" >/dev/null
    done
    end_ns="$(now_ns)"
    duration_ms="$(( (end_ns - start_ns) / 1000000 ))"
    record_timing "sign_total" "${workload}" "${iteration}" "${duration_ms}"

    start_ns="$(now_ns)"
    "${BIN}" verify \
      --source local \
      --attestations "${bundle_dir}" \
      --policy policy/examples/mvp-gates.yaml \
      --format json \
      --out "${work_dir}/verify-${iteration}.json" >/dev/null
    end_ns="$(now_ns)"
    duration_ms="$(( (end_ns - start_ns) / 1000000 ))"
    record_timing "verify_total" "${workload}" "${iteration}" "${duration_ms}"
  done
}

measure_policy_scale() {
  local evaluations=100
  for iteration in 1 2 3; do
    local start_ns end_ns duration_ms
    start_ns="$(now_ns)"
    for _ in $(seq 1 "${evaluations}"); do
      "${BIN}" gate \
        --policy policy/examples/mvp-gates.yaml \
        --attestations .llmsa/attestations \
        --git-ref HEAD >/dev/null
    done
    end_ns="$(now_ns)"
    duration_ms="$(( (end_ns - start_ns) / 1000000 ))"
    record_timing "policy_total" "${evaluations}" "${iteration}" "${duration_ms}"
  done
}

generate_reproducibility_metrics() {
  local repro_dir="${RUN_DIR}/repro"
  mkdir -p "${repro_dir}"
  "${BIN}" attest create --type prompt_attestation --config examples/tiny-rag/configs/prompt.yaml --out "${repro_dir}" --determinism-check 2 >/dev/null
  "${BIN}" attest create --type prompt_attestation --config examples/tiny-rag/configs/prompt.yaml --out "${repro_dir}" --determinism-check 2 >/dev/null

  python3 - "${repro_dir}" "${REPRO_JSON}" <<'PY'
import glob
import hashlib
import json
import os
import sys

repro_dir, out_file = sys.argv[1], sys.argv[2]
files = sorted(glob.glob(os.path.join(repro_dir, "statement_prompt_attestation_*.json")))
if len(files) < 2:
    raise SystemExit("need at least two statements for reproducibility check")

def normalized_hash(path):
    with open(path, "r", encoding="utf-8") as f:
        data = json.load(f)
    data["statement_id"] = "__normalized__"
    data["generated_at"] = "__normalized__"
    raw = json.dumps(data, sort_keys=True, separators=(",", ":")).encode("utf-8")
    return hashlib.sha256(raw).hexdigest()

h1 = normalized_hash(files[-2])
h2 = normalized_hash(files[-1])
result = {
    "statement_a": files[-2],
    "statement_b": files[-1],
    "normalized_hash_a": h1,
    "normalized_hash_b": h2,
    "stable": h1 == h2,
    "stability_rate": 1.0 if h1 == h2 else 0.0
}
with open(out_file, "w", encoding="utf-8") as f:
    json.dump(result, f, indent=2)
PY
}

write_manifest() {
  python3 - "${MANIFEST_JSON}" <<'PY'
import json
import os
import platform
import subprocess
import sys
from datetime import datetime, timezone

out_file = sys.argv[1]

def cmd(*args):
    try:
        return subprocess.check_output(args, text=True).strip()
    except Exception:
        return ""

manifest = {
    "generated_at_utc": datetime.now(timezone.utc).isoformat(),
    "git_sha": cmd("git", "rev-parse", "HEAD"),
    "git_status_porcelain": cmd("git", "status", "--short"),
    "go_version": cmd("go", "version"),
    "platform": platform.platform(),
    "machine": platform.machine(),
    "python_version": platform.python_version(),
}
with open(out_file, "w", encoding="utf-8") as f:
    json.dump(manifest, f, indent=2)
PY
}

render_summary() {
  python3 - "${CSV_FILE}" "${REPRO_JSON}" "${SUMMARY_MD}" "${DOCS_FILE}" "${RUN_DIR}" <<'PY'
import csv
import json
import statistics
import sys
from collections import defaultdict

csv_file, repro_file, summary_file, docs_file, run_dir = sys.argv[1:6]

rows = list(csv.DictReader(open(csv_file, "r", encoding="utf-8")))
repro = json.load(open(repro_file, "r", encoding="utf-8"))

by_key = defaultdict(list)
for row in rows:
    key = (row["scenario"], row["workload"])
    by_key[key].append(float(row["duration_ms"]))

lines = [
    "# LLM Supply-Chain Attestation Benchmark Summary",
    "",
    f"- Run Directory: `{run_dir}`",
    f"- Raw Timing CSV: `{csv_file}`",
    f"- Reproducibility JSON: `{repro_file}`",
    "",
    "## Timing",
    "",
    "| Scenario | Workload | Samples | p50 ms | p95 ms |",
    "|---|---:|---:|---:|---:|",
]

for key in sorted(by_key):
    values = sorted(by_key[key])
    p50 = statistics.median(values)
    idx95 = max(0, int(round(0.95 * len(values))) - 1)
    p95 = values[idx95]
    lines.append(f"| {key[0]} | {key[1]} | {len(values)} | {p50:.1f} | {p95:.1f} |")

lines.extend(
    [
        "",
        "## Reproducibility",
        "",
        f"- Stable: `{repro.get('stable')}`",
        f"- Stability Rate: `{repro.get('stability_rate')}`",
        "",
        "## Limitations",
        "",
        "- Workloads are synthetic copies of benchmark fixtures, not production-scale corpora.",
        "- Signing uses local PEM-backed sigstore mode for repeatable offline timing.",
        "- Results should be compared by trend over nightly runs, not a single run.",
        "",
    ]
)

content = "\n".join(lines)
open(summary_file, "w", encoding="utf-8").write(content)
open(docs_file, "w", encoding="utf-8").write(content)
PY
}

go build -o "${BIN}" ./cmd/llmsa
make -C examples/tiny-rag clean attest sign >/dev/null

BASE_STATEMENT="$(ls .llmsa/attestations/statement_prompt_attestation_*.json | head -n1)"
measure_sign_and_verify 1 "${BASE_STATEMENT}"
measure_sign_and_verify 10 "${BASE_STATEMENT}"
measure_sign_and_verify 100 "${BASE_STATEMENT}"
measure_policy_scale
generate_reproducibility_metrics
write_manifest
render_summary

echo "benchmark artifacts generated in ${RUN_DIR}"
