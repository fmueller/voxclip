#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
zero="0000000000000000000000000000000000000000"

collect_ranges() {
  if [ "$#" -gt 0 ]; then
    printf '%s\n' "$@"
    return
  fi
  if [ ! -t 0 ]; then
    while read -r _local_ref local_sha _remote_ref remote_sha; do
      [ -z "${local_sha:-}" ] && continue
      [ "$local_sha" = "$zero" ] && continue
      if [ "${remote_sha:-$zero}" = "$zero" ] || ! git cat-file -e "$remote_sha^{commit}" 2>/dev/null; then
        printf '%s\n' "$local_sha"
      else
        printf '%s..%s\n' "$remote_sha" "$local_sha"
      fi
    done
    return
  fi
  if upstream="$(git rev-parse --abbrev-ref --symbolic-full-name '@{upstream}' 2>/dev/null)"; then
    printf '%s..HEAD\n' "$upstream"
  else
    printf 'HEAD\n'
  fi
}

ranges="$(collect_ranges "$@")"
[ -z "$ranges" ] && exit 0

commits="$(printf '%s\n' "$ranges" | while read -r range; do
  [ -z "$range" ] && continue
  git rev-list --max-count=200 "$range"
done | sort -u)"
[ -z "$commits" ] && exit 0

message_file="$(mktemp)"
trap 'rm -f "$message_file"' EXIT
status=0
for commit in $commits; do
  git log -1 --format='%B' "$commit" >"$message_file"
  if ! bash "$script_dir/check-commit-msg.sh" "$message_file" >/dev/null; then
    echo "check-push-messages: $commit violates the commit message policy" >&2
    git log -1 --format='  %h %s' "$commit" >&2
    status=1
  fi
  if ! bash "$script_dir/check-author.sh" "$(git log -1 --format='%an <%ae>' "$commit")" >/dev/null; then
    echo "check-push-messages: $commit was authored by an agent identity" >&2
    git log -1 --format='  %h %an <%ae>' "$commit" >&2
    status=1
  fi
done

if [ "$status" -ne 0 ]; then
  echo "check-push-messages: rewrite those commits before pushing" >&2
fi
exit "$status"
