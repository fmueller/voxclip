#!/usr/bin/env bash
set -euo pipefail

msg_file="${1:-}"
if [ -z "$msg_file" ] || [ ! -f "$msg_file" ]; then
  echo "check-commit-msg: missing commit message file argument" >&2
  exit 1
fi

subject="$(grep -vE '^[[:space:]]*#' "$msg_file" | sed '/^[[:space:]]*$/d' | head -n 1)"

case "$subject" in
  "Merge "* | "Revert "* | "fixup! "* | "squash! "*)
    ;;
  *)
    if ! printf '%s' "$subject" | grep -qE '^(feat|fix|refactor|docs|test|chore|build|perf|ci)(\([a-z0-9._-]+\))?!?: .+$'; then
      echo "check-commit-msg: subject must be a Conventional Commit:" >&2
      echo "  <type>: <description>" >&2
      echo "  types: feat fix refactor docs test chore build perf ci" >&2
      echo "got: ${subject:-<empty>}" >&2
      exit 1
    fi
    ;;
esac

if ! bash "$(dirname "${BASH_SOURCE[0]}")/check-attribution.sh" "$msg_file"; then
  echo "check-commit-msg: remove automated attribution" >&2
  exit 1
fi
