#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
checker="$script_dir/check-commit-msg.sh"
tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

fail() {
  printf 'FAIL: %s\n' "$*" >&2
  exit 1
}

assert_accepts() {
  local name="$1"
  local message="$2"
  local message_file="$tmp_dir/$name"
  local output

  printf '%s\n' "$message" >"$message_file"
  if ! output="$(bash "$checker" "$message_file" 2>&1)"; then
    fail "$name was rejected: $output"
  fi
}

assert_rejects() {
  local name="$1"
  local message="$2"
  local expected="$3"
  local message_file="$tmp_dir/$name"
  local output

  printf '%s\n' "$message" >"$message_file"
  if output="$(bash "$checker" "$message_file" 2>&1)"; then
    fail "$name was accepted"
  fi
  if [[ "$output" != *"$expected"* ]]; then
    fail "$name did not report '$expected': $output"
  fi
}

assert_accepts conventional 'chore: configure repository hooks'
assert_accepts scoped 'fix(cli): preserve transcript output'
assert_accepts breaking 'feat!: change the default model'
assert_accepts generated-merge "Merge branch 'feature'"
assert_accepts generated-revert 'Revert "feat: change the default model"'

assert_rejects empty '' 'Conventional Commit'
assert_rejects invalid-type 'setup: configure repository hooks' 'Conventional Commit'
assert_rejects invalid-subject 'configure repository hooks' 'Conventional Commit'
assert_rejects attribution $'chore: configure repository hooks\n\nCo-authored-by: Bot <bot@example.com>' 'automated attribution'
assert_rejects thread-link $'chore: configure repository hooks\n\nhttps://ampcode.com/threads/T-1' 'automated attribution'

printf 'commit message checks passed\n'
