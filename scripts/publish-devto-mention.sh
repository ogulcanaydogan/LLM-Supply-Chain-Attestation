#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

if ! command -v curl >/dev/null 2>&1; then
  echo "error: curl is required" >&2
  exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
  echo "error: jq is required" >&2
  exit 1
fi

if [[ -z "${DEVTO_API_KEY:-}" ]]; then
  echo "error: DEVTO_API_KEY is not set" >&2
  echo "hint: export DEVTO_API_KEY=<key> and rerun" >&2
  exit 1
fi

TITLE="${1:-From Scripts to Evidence: LLM Artifact Integrity in CI/CD}"
DRAFT_FILE="${2:-docs/public-footprint/third-party-mention-draft-2026-02.md}"
PUBLISHED_FLAG="${PUBLISHED_FLAG:-false}"

if [[ ! -f "${DRAFT_FILE}" ]]; then
  echo "error: draft file not found: ${DRAFT_FILE}" >&2
  exit 1
fi

BODY="$(cat "${DRAFT_FILE}")"
PAYLOAD="$(jq -n \
  --arg title "${TITLE}" \
  --arg body "${BODY}" \
  --argjson published "${PUBLISHED_FLAG}" \
  '{
    article: {
      title: $title,
      body_markdown: $body,
      published: $published,
      tags: ["security", "devops", "supplychain", "opa", "sigstore"]
    }
  }')"

RESPONSE="$(curl -sS -X POST "https://dev.to/api/articles" \
  -H "api-key: ${DEVTO_API_KEY}" \
  -H "Content-Type: application/json" \
  -d "${PAYLOAD}")"

URL="$(echo "${RESPONSE}" | jq -r '.url // empty')"
if [[ -z "${URL}" ]]; then
  echo "error: failed to publish to dev.to" >&2
  echo "${RESPONSE}" | jq . >&2 || echo "${RESPONSE}" >&2
  exit 1
fi

echo "published_url=${URL}"
