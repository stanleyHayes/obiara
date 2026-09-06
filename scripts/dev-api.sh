#!/usr/bin/env bash
#
# Runs the API locally with real delivery providers.
#
# The service reads os.Getenv directly and loads no env file (secrets arrive
# only through the environment; services/api/internal/platform/config). So the
# file has to be put into the environment by something, and this is it.
#
# Without this the API falls back to configuration defaults, which means the
# OTP simulator — and the simulator delivers nothing and deliberately never
# logs the code, because codes are secrets (agent_plan.md §14). A local
# sign-in could not receive a code by any route.
#
# Usage: scripts/dev-api.sh
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
env_file="$root/services/api/.env.development.local"

if [[ ! -f "$env_file" ]]; then
  echo "missing $env_file" >&2
  echo "Copy services/api/.env.production to it, then set APP_ENV=development" >&2
  echo "and MONGODB_DATABASE=obiara_dev." >&2
  exit 1
fi

# Read as data, never evaluated: values hold & and spaces (the Mongo URI and
# the SMS template both do), and sourcing the file would let the shell act on
# them.
while IFS= read -r line || [[ -n "$line" ]]; do
  [[ -z "$line" || "$line" == \#* ]] && continue
  [[ "$line" != *=* ]] && continue
  export "${line%%=*}=${line#*=}"
done <"$env_file"

# A laptop must never come up as production. The env file sets this, and this
# is the check that it was not edited back.
if [[ "${APP_ENV:-}" != "development" ]]; then
  echo "refusing to start: APP_ENV is '${APP_ENV:-unset}', expected development" >&2
  exit 1
fi
if [[ "${MONGODB_DATABASE:-}" == *production* ]]; then
  echo "refusing to start: MONGODB_DATABASE is '${MONGODB_DATABASE}'" >&2
  exit 1
fi

echo "api: APP_ENV=$APP_ENV db=$MONGODB_DATABASE otp=${OTP_PROVIDERS:-simulator} email=${EMAIL_PROVIDER:-simulator}"
cd "$root"
exec go run ./services/api
