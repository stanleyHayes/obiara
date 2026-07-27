#!/usr/bin/env bash
set -euo pipefail

output_path="${1:-release-evidence.json}"
target="${RELEASE_TARGET:?RELEASE_TARGET is required}"
candidate="${RELEASE_COMMIT_SHA:?RELEASE_COMMIT_SHA is required}"
change_record="${RELEASE_CHANGE_RECORD:?RELEASE_CHANGE_RECORD is required}"
repository="${RELEASE_REPOSITORY:?RELEASE_REPOSITORY is required}"

case "$target" in
  preview|staging|production) ;;
  *) echo "unsupported release target" >&2; exit 2 ;;
esac

if [[ ! "$candidate" =~ ^[0-9a-f]{40}$ ]]; then
  echo "release commit must be a lowercase 40-character Git SHA" >&2
  exit 2
fi
if [[ ! "$change_record" =~ ^[A-Za-z0-9][A-Za-z0-9._:/-]{2,79}$ ]]; then
  echo "change record must be a bounded opaque reference" >&2
  exit 2
fi
if [[ "$(git rev-parse HEAD)" != "$candidate" ]]; then
  echo "checked-out commit does not match the requested candidate" >&2
  exit 3
fi

git fetch --no-tags origin main
if [[ "$(git rev-parse origin/main)" != "$candidate" ]]; then
  echo "candidate is not the current immutable main commit" >&2
  exit 3
fi

required_checks=(
  "Lint, test, and build"
  "CodeQL (go)"
  "CodeQL (javascript-typescript)"
  "Dependency vulnerabilities"
)
checks_json="$(gh api \
  -H "Accept: application/vnd.github+json" \
  "repos/${repository}/commits/${candidate}/check-runs?per_page=100")"

for check_name in "${required_checks[@]}"; do
  conclusion="$(jq -r --arg name "$check_name" '
    [.check_runs[] | select(.name == $name) | .conclusion] | last // "missing"
  ' <<<"$checks_json")"
  if [[ "$conclusion" != "success" ]]; then
    echo "required check is not successful: ${check_name}" >&2
    exit 4
  fi
done

if [[ "$target" == "production" ]]; then
  echo "production is blocked: approved topology and signed release gates are absent" >&2
  exit 5
fi

jq -n \
  --arg schemaVersion "obiara.release-evidence.v1" \
  --arg target "$target" \
  --arg commitSha "$candidate" \
  --arg repository "$repository" \
  --arg sourceRef "${RELEASE_SOURCE_REF:-unknown}" \
  --arg workflowRun "${RELEASE_RUN_ID:-local}" \
  --arg requestedBy "${RELEASE_REQUESTED_BY:-unknown}" \
  --arg changeRecord "$change_record" \
  --arg generatedAt "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  --argjson requiredChecks "$(printf '%s\n' "${required_checks[@]}" | jq -R . | jq -s .)" \
  '{
    schemaVersion: $schemaVersion,
    target: $target,
    commitSha: $commitSha,
    repository: $repository,
    sourceRef: $sourceRef,
    workflowRun: $workflowRun,
    requestedBy: $requestedBy,
    changeRecord: $changeRecord,
    generatedAt: $generatedAt,
    requiredChecks: $requiredChecks,
    disposition: "qualified-non-production"
  }' > "$output_path"

