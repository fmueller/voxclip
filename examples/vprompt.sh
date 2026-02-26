#!/bin/sh
# vprompt.sh — capture a voice prompt as text via command substitution.
#
# Source this file, then pass $(vprompt) to your agent CLI.
# Works with terminal tools like claude, codex, opencode, crush, and aider.
#
# Configure recording duration (default 10s):
#   export VPROMPT_DURATION=15s

vprompt() {
  voxclip --no-progress --duration "${VPROMPT_DURATION:-10s}" 2>/dev/null || {
    echo "vprompt: recording failed; is voxclip installed?" >&2
    return 1
  }
}
