#!/usr/bin/env bash
set -euo pipefail

msg_file="${1:-}"
if [ -z "$msg_file" ] || [ ! -f "$msg_file" ]; then
  echo "check-attribution: missing commit message file argument" >&2
  exit 2
fi

if grep -qiE '^[[:space:]]*((co-authored-by|assisted-by|generated-by|(claude|amp|agent|codex|copilot)-(session|thread)(-id)?):|generated with[[:space:]])' "$msg_file" \
  || grep -qiE '(claude\.ai/code/session|ampcode\.com/threads|chatgpt\.com/(share|c)/|cursor\.com/agents)' "$msg_file" \
  || grep -qF '🤖' "$msg_file"; then
  exit 1
fi
