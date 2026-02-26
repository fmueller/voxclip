#!/bin/sh
# vprompt-interactive.sh — interactive voice capture with immediate start.
#
# Source this file, then use it while your coding agent session is already open:
#   vprompt_interactive
#   # Speak, press Enter when done. Transcript is on your clipboard.
#   # Paste with Cmd+V (macOS) or Ctrl+Shift+V (terminal).

vprompt_interactive() {
  voxclip --immediate 2>/dev/null || {
    echo "vprompt_interactive: recording failed; is voxclip installed?" >&2
    return 1
  }
}
