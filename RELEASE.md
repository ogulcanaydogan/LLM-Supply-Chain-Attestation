# Release

## MVP release checklist
- `go test ./...` passes.
- CI `attestation-gate` passes.
- Verify report artifacts published (`verify.json`, `verify.md`).

## release-verify auth troubleshooting

The `release-verify` workflow runs a **preflight** step ("Validate GitHub API
access") before downloading release assets. It probes the workflow token, the
GitHub API, and the resolved release so that auth problems fail fast with an
actionable message instead of surfacing as opaque "failed to download release
assets after retries" errors three retries later.

Common failure modes and remediation:

| Preflight error | Likely cause | Remediation |
| --- | --- | --- |
| `GH_TOKEN is empty` | Job lost `permissions: contents: read`, or `GH_TOKEN` not wired to `${{ github.token }}` | Restore `permissions: contents: read` on the job and re-run. |
| `gh CLI is not authenticated` | `GITHUB_TOKEN` expired/revoked or missing `contents: read` scope | Run `gh auth status` locally with the same token, confirm workflow permissions, re-run. |
| `GitHub API request failed for <repo>` | Token lacks repo read access, transient API outage, or secondary rate limiting | Verify `contents: read`; if rate limited, wait and re-run. |
| `release '<tag>' is not readable` | Tag does not exist, release not published yet, or insufficient read scope | `gh release list -R <repo>` to confirm, re-run with a valid published tag. |

Local debugging:

```bash
# Confirm the token the workflow uses can see the release
gh auth status
gh api "/repos/<owner>/<repo>" --jq '.full_name'
gh release view "<tag>" -R "<owner>/<repo>"
```

The preflight only validates access; the real integrity gates (`checksums.txt`
verification and `cosign verify-blob` signature/identity checks) still run and
still fail the workflow on any mismatch.
