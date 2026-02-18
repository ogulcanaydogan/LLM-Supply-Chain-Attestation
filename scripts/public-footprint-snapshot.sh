#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

if ! command -v gh >/dev/null 2>&1; then
  echo "error: gh CLI is required" >&2
  exit 1
fi

if ! command -v python3 >/dev/null 2>&1; then
  echo "error: python3 is required" >&2
  exit 1
fi

if ! gh auth status >/dev/null 2>&1; then
  if [[ -z "${GH_TOKEN:-}" && -z "${GITHUB_TOKEN:-}" ]]; then
    echo "error: gh is not authenticated. run: gh auth login" >&2
    exit 1
  fi
fi

REPO="${1:-}"
if [[ -z "${REPO}" ]]; then
  REPO="$(gh repo view --json nameWithOwner -q .nameWithOwner)"
fi

TS="$(date -u +"%Y%m%dT%H%M%SZ")"
OUT_DIR=".llmsa/public-footprint/${TS}"
mkdir -p "${OUT_DIR}"

REPO_JSON="${OUT_DIR}/repo.json"
RELEASES_JSON="${OUT_DIR}/releases.json"
RUNS_JSON="${OUT_DIR}/runs.json"
PRS_JSON="${OUT_DIR}/prs.json"
SNAPSHOT_JSON="${OUT_DIR}/snapshot.json"
SNAPSHOT_MD="${OUT_DIR}/snapshot.md"

gh repo view "${REPO}" --json nameWithOwner,url,stargazerCount,forkCount,watchers,defaultBranchRef,updatedAt,createdAt > "${REPO_JSON}"
gh api "repos/${REPO}/releases?per_page=100" > "${RELEASES_JSON}"
gh run list --repo "${REPO}" --limit 200 --json conclusion,createdAt,name,url > "${RUNS_JSON}"
gh pr list --repo "${REPO}" --state all --limit 200 --json createdAt,mergedAt,state,url > "${PRS_JSON}"

python3 - "${REPO_JSON}" "${RELEASES_JSON}" "${RUNS_JSON}" "${PRS_JSON}" "${SNAPSHOT_JSON}" "${SNAPSHOT_MD}" <<'PY'
import json
import sys
from datetime import datetime, timedelta, timezone

repo_file, releases_file, runs_file, prs_file, out_json, out_md = sys.argv[1:7]

def parse_time(value):
    try:
        return datetime.fromisoformat(value.replace("Z", "+00:00"))
    except Exception:
        return None

now = datetime.now(timezone.utc)
window_start = now - timedelta(days=30)

repo = json.load(open(repo_file, "r", encoding="utf-8"))
releases = json.load(open(releases_file, "r", encoding="utf-8"))
runs = json.load(open(runs_file, "r", encoding="utf-8"))
prs = json.load(open(prs_file, "r", encoding="utf-8"))

downloads_total = 0
for release in releases:
    for asset in release.get("assets", []):
        downloads_total += int(asset.get("download_count", 0))

runs_30d = []
for run in runs:
    ts = parse_time(run.get("createdAt", ""))
    if ts and ts >= window_start:
        runs_30d.append(run)

success_30d = sum(1 for run in runs_30d if run.get("conclusion") == "success")
total_30d = len(runs_30d)
pass_rate_30d = round((success_30d / total_30d) * 100.0, 2) if total_30d else 0.0

prs_30d = []
for pr in prs:
    ts = parse_time(pr.get("createdAt", ""))
    if ts and ts >= window_start:
        prs_30d.append(pr)

opened_prs_30d = len(prs_30d)
merged_prs_30d = sum(1 for pr in prs_30d if pr.get("mergedAt"))

snapshot = {
    "generated_at_utc": now.isoformat(),
    "window_days": 30,
    "repo": repo.get("nameWithOwner"),
    "repo_url": repo.get("url"),
    "stars": repo.get("stargazerCount", 0),
    "forks": repo.get("forkCount", 0),
    "watchers": repo.get("watchers", {}).get("totalCount", 0),
    "default_branch": repo.get("defaultBranchRef", {}).get("name"),
    "release_count": len(releases),
    "release_downloads_total": downloads_total,
    "ci_runs_30d_total": total_30d,
    "ci_runs_30d_success": success_30d,
    "ci_pass_rate_30d_percent": pass_rate_30d,
    "prs_opened_30d": opened_prs_30d,
    "prs_merged_30d": merged_prs_30d,
}

with open(out_json, "w", encoding="utf-8") as f:
    json.dump(snapshot, f, indent=2, sort_keys=True)

lines = [
    "# Public Footprint Snapshot",
    "",
    f"- Generated (UTC): `{snapshot['generated_at_utc']}`",
    f"- Repository: `{snapshot['repo']}`",
    "",
    "## Metrics",
    "",
    "| Metric | Value |",
    "|---|---:|",
    f"| Stars | {snapshot['stars']} |",
    f"| Forks | {snapshot['forks']} |",
    f"| Watchers | {snapshot['watchers']} |",
    f"| Releases | {snapshot['release_count']} |",
    f"| Release downloads (cumulative) | {snapshot['release_downloads_total']} |",
    f"| CI runs (last 30 days) | {snapshot['ci_runs_30d_total']} |",
    f"| CI pass rate (last 30 days) | {snapshot['ci_pass_rate_30d_percent']}% |",
    f"| PRs opened (last 30 days) | {snapshot['prs_opened_30d']} |",
    f"| PRs merged (last 30 days) | {snapshot['prs_merged_30d']} |",
]

with open(out_md, "w", encoding="utf-8") as f:
    f.write("\n".join(lines) + "\n")
PY

echo "snapshot json: ${SNAPSHOT_JSON}"
echo "snapshot markdown: ${SNAPSHOT_MD}"
