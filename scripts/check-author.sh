#!/usr/bin/env bash
set -euo pipefail

identity="${1:-}"
if [ -z "$identity" ]; then
  identity="$(git var GIT_AUTHOR_IDENT | sed -E 's/> [0-9]+ [+-][0-9]{4}$/>/')"
fi

if printf '%s' "$identity" | grep -qiE '<[^>]*@(ampcode\.com|anthropic\.com|openai\.com|cursor\.(com|sh)|devin\.ai)>' \
  || printf '%s' "$identity" | grep -qiE '^(amp|claude|codex|copilot|cursor|devin|agent|bot)([[:space:]]|<|-)' \
  || printf '%s' "$identity" | grep -qiE '<(amp|claude|codex|copilot|cursor|devin)(-?agent)?@'; then
  echo "check-author: refusing a commit authored by an agent identity:" >&2
  echo "  $identity" >&2
  echo "set user.name and user.email to a person before committing" >&2
  exit 1
fi
