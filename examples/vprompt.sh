#!/bin/sh
# vprompt.sh — capture a voice prompt as text via command substitution.
#
# Source this file, then use:
#   claude "$(vprompt)"
#   aider --message "$(vprompt)"
#
# Configure recording duration (default 10s):
#   export VPROMPT_DURATION=15s

vprompt() {
  voxclip --no-progress --duration "${VPROMPT_DURATION:-10s}" 2>/dev/null
}
