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

repo_from_git_remote() {
  local remote
  remote="$(git config --get remote.origin.url 2>/dev/null || true)"
  case "${remote}" in
    git@github.com:*)
      echo "${remote#git@github.com:}" | sed 's/\.git$//'
      ;;
    https://github.com/*)
      echo "${remote#https://github.com/}" | sed 's/\.git$//'
      ;;
    http://github.com/*)
      echo "${remote#http://github.com/}" | sed 's/\.git$//'
      ;;
    *)
      echo ""
      ;;
  esac
}

gh_api_retry() {
  local attempt
  for attempt in 1 2 3; do
    if gh api "$@"; then
      return 0
    fi
    sleep $((attempt * 2))
  done
  return 1
}

extract_upstream_pr_urls() {
  local source_file="$1"
  if [[ ! -f "${source_file}" ]]; then
    return
  fi
  grep -Eo 'https://github\.com/[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+/pull/[0-9]+' "${source_file}" | sort -u || true
}

parse_pr_repo_number() {
  local pr_url="$1"
  if [[ "${pr_url}" =~ ^https://github\.com/([^/]+)/([^/]+)/pull/([0-9]+)$ ]]; then
    echo "${BASH_REMATCH[1]}/${BASH_REMATCH[2]}:${BASH_REMATCH[3]}"
    return
  fi
  echo ""
}

if ! gh auth status >/dev/null 2>&1; then
  if [[ -z "${GH_TOKEN:-}" && -z "${GITHUB_TOKEN:-}" ]]; then
    echo "warning: gh is not authenticated; using anonymous API access (rate-limited)." >&2
  fi
fi

REPO="${1:-}"
if [[ -z "${REPO}" ]]; then
  REPO="$(repo_from_git_remote)"
fi
if [[ -z "${REPO}" ]]; then
  REPO="$(gh api user/repos --jq '.[0].full_name' 2>/dev/null || true)"
fi
if [[ -z "${REPO}" || "${REPO}" == "null" ]]; then
  echo "error: repository argument is required when gh repo context is unavailable" >&2
  exit 1
fi

TS="$(date -u +"%Y%m%dT%H%M%SZ")"
OUT_DIR=".llmsa/public-footprint/${TS}"
mkdir -p "${OUT_DIR}"

EXTERNAL_LOG="docs/public-footprint/external-contribution-log.md"
REPO_JSON="${OUT_DIR}/repo.json"
RELEASES_JSON="${OUT_DIR}/releases.json"
RUNS_JSON="${OUT_DIR}/runs.json"
PRS_JSON="${OUT_DIR}/prs.json"
UPSTREAM_PRS_JSON="${OUT_DIR}/upstream-prs.json"
SNAPSHOT_JSON="${OUT_DIR}/snapshot.json"
SNAPSHOT_MD="${OUT_DIR}/snapshot.md"

gh_api_retry "repos/${REPO}" > "${REPO_JSON}"
gh_api_retry "repos/${REPO}/releases?per_page=100" > "${RELEASES_JSON}"
gh_api_retry "repos/${REPO}/actions/runs?per_page=100" \
  --jq '[.workflow_runs[] | {conclusion: .conclusion, createdAt: .created_at, name: .name, url: .html_url}]' > "${RUNS_JSON}"
gh_api_retry "repos/${REPO}/pulls?state=all&per_page=100" \
  --jq '[.[] | {createdAt: .created_at, mergedAt: .merged_at, state: .state, url: .html_url}]' > "${PRS_JSON}"

tmp_upstream_prs_ndjson="${OUT_DIR}/upstream-prs.ndjson"
: > "${tmp_upstream_prs_ndjson}"

mapfile -t upstream_pr_urls < <(extract_upstream_pr_urls "${EXTERNAL_LOG}")
if [[ "${#upstream_pr_urls[@]}" -eq 0 ]]; then
  upstream_pr_urls=(
    "https://github.com/sigstore/cosign/pull/4710"
    "https://github.com/open-policy-agent/opa/pull/8343"
    "https://github.com/ossf/scorecard/pull/4942"
  )
fi

for pr_url in "${upstream_pr_urls[@]}"; do
  repo_number="$(parse_pr_repo_number "${pr_url}")"
  if [[ -z "${repo_number}" ]]; then
    continue
  fi
  pr_repo="${repo_number%%:*}"
  pr_number="${repo_number##*:}"
  gh_api_retry "repos/${pr_repo}/pulls/${pr_number}" \
    --jq '{html_url: .html_url, state: .state, merged: .merged, merged_at: .merged_at, draft: .draft}' \
    >> "${tmp_upstream_prs_ndjson}"
done

if [[ -s "${tmp_upstream_prs_ndjson}" ]]; then
  jq -s '.' "${tmp_upstream_prs_ndjson}" > "${UPSTREAM_PRS_JSON}"
else
  echo '[]' > "${UPSTREAM_PRS_JSON}"
fi

python3 - "${REPO_JSON}" "${RELEASES_JSON}" "${RUNS_JSON}" "${PRS_JSON}" "${UPSTREAM_PRS_JSON}" "${SNAPSHOT_JSON}" "${SNAPSHOT_MD}" <<'PY'
import json
import sys
from datetime import datetime, timedelta, timezone

(
    repo_file,
    releases_file,
    runs_file,
    prs_file,
    upstream_prs_file,
    out_json,
    out_md,
) = sys.argv[1:8]

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
upstream_prs = json.load(open(upstream_prs_file, "r", encoding="utf-8"))

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

upstream_pr_open = sum(1 for pr in upstream_prs if pr.get("state") == "open")
upstream_pr_merged = sum(1 for pr in upstream_prs if pr.get("merged"))
upstream_pr_in_review = sum(
    1
    for pr in upstream_prs
    if pr.get("state") == "open" and not pr.get("draft") and not pr.get("merged")
)
upstream_pr_closed_unmerged = sum(
    1 for pr in upstream_prs if pr.get("state") == "closed" and not pr.get("merged")
)

watchers_raw = repo.get("watchers")
watchers = 0
if isinstance(watchers_raw, dict):
    watchers = watchers_raw.get("totalCount", 0)
elif isinstance(watchers_raw, int):
    watchers = watchers_raw
if not watchers:
    watchers = repo.get("subscribers_count", repo.get("watchers_count", 0))

default_branch_ref = repo.get("defaultBranchRef")
default_branch = repo.get("default_branch")
if isinstance(default_branch_ref, dict) and default_branch_ref.get("name"):
    default_branch = default_branch_ref.get("name")

snapshot = {
    "generated_at_utc": now.isoformat(),
    "window_days": 30,
    "repo": repo.get("nameWithOwner") or repo.get("full_name"),
    "repo_url": repo.get("url") or repo.get("html_url"),
    "stars": repo.get("stargazerCount", repo.get("stargazers_count", 0)),
    "forks": repo.get("forkCount", repo.get("forks_count", 0)),
    "watchers": watchers,
    "default_branch": default_branch,
    "release_count": len(releases),
    "release_downloads_total": downloads_total,
    "ci_runs_30d_total": total_30d,
    "ci_runs_30d_success": success_30d,
    "ci_pass_rate_30d_percent": pass_rate_30d,
    "prs_opened_30d": opened_prs_30d,
    "prs_merged_30d": merged_prs_30d,
    "upstream_pr_open": upstream_pr_open,
    "upstream_pr_merged": upstream_pr_merged,
    "in_review_count": upstream_pr_in_review,
    "upstream_pr_closed_unmerged": upstream_pr_closed_unmerged,
    "upstream_pr_urls": [pr.get("html_url") for pr in upstream_prs if pr.get("html_url")],
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
    f"| Upstream PRs open | {snapshot['upstream_pr_open']} |",
    f"| Upstream PRs merged | {snapshot['upstream_pr_merged']} |",
    f"| Upstream PRs in review | {snapshot['in_review_count']} |",
    f"| Upstream PRs closed (unmerged) | {snapshot['upstream_pr_closed_unmerged']} |",
]

with open(out_md, "w", encoding="utf-8") as f:
    f.write("\n".join(lines) + "\n")
PY

echo "snapshot json: ${SNAPSHOT_JSON}"
echo "snapshot markdown: ${SNAPSHOT_MD}"
