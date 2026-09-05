#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
scanner="$script_dir/check-push-messages.sh"
tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

fail() {
  printf 'FAIL: %s\n' "$*" >&2
  exit 1
}

repo="$tmp_dir/repo"
git init -q "$repo"
cd "$repo"
git config user.email fixture@example.com
git config user.name Fixture
git config commit.gpgsign false
mkdir "$repo/hooks"
git config core.hooksPath "$repo/hooks"

commit() {
  printf '%s\n' "$2" >>file.txt
  git add file.txt
  git commit -q --no-verify -m "$1"
}

commit 'feat: add the first fixture' one
clean_head="$(git rev-parse HEAD)"
if ! output="$(bash "$scanner" "$clean_head" 2>&1)"; then
  fail "a clean commit was rejected: $output"
fi

commit 'add the second fixture' two
if output="$(bash "$scanner" "$clean_head..HEAD" 2>&1)"; then
  fail "a non-Conventional Commit was accepted"
fi
if [[ "$output" != *"commit message policy"* ]]; then
  fail "the scanner did not name the message policy: $output"
fi
git reset -q --hard "$clean_head"

commit $'feat: add the second fixture\n\nAmp-Thread: https://ampcode.com/threads/T-1' two
if output="$(bash "$scanner" "$clean_head..HEAD" 2>&1)"; then
  fail "an attribution trailer was accepted"
fi
if [[ "$output" != *"automated attribution"* ]]; then
  fail "the scanner did not name the attribution policy: $output"
fi

if printf 'refs/heads/main %s refs/heads/main %s\n' "$(git rev-parse HEAD)" "$clean_head" | bash "$scanner" >/dev/null 2>&1; then
  fail "the stdin range accepted an attribution trailer"
fi
if ! printf 'refs/heads/main %s refs/heads/main %s\n' "$clean_head" "$clean_head" | bash "$scanner" >/dev/null 2>&1; then
  fail "the stdin range rejected an empty push"
fi

commit 'feat: add the third fixture' three
agent_parent="$(git rev-parse HEAD~1)"
git commit -q --amend --no-verify --no-edit --author='Amp <amp@ampcode.com>'
if output="$(bash "$scanner" "$agent_parent..HEAD" 2>&1)"; then
  fail "an agent author was accepted"
fi
if [[ "$output" != *"agent identity"* ]]; then
  fail "the scanner did not name the author policy: $output"
fi

printf 'push message checks passed\n'
