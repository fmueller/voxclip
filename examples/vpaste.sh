#!/bin/sh
# vpaste.sh — voice-to-paste: record, transcribe, paste into active window.
#
# macOS: uses osascript to simulate Cmd+V
# Linux: uses xdotool to simulate Ctrl+Shift+V (terminal-friendly)
#
# Usage:
#   Bind this script to a global hotkey (see examples/README.md).
#   Press the hotkey, speak for N seconds, transcript auto-pastes.
#
# Configure recording duration (default 8s):
#   export VPROMPT_DURATION=12s

set -e

export PATH="$HOME/.local/bin:/opt/homebrew/bin:/usr/local/bin:$PATH"

voxclip --duration "${VPROMPT_DURATION:-8s}" --no-progress 2>/dev/null

case "$(uname -s)" in
  Darwin)
    osascript -e 'delay 0.2' -e 'tell application "System Events" to keystroke "v" using command down'
    ;;
  Linux)
    sleep 0.2
    if command -v xdotool >/dev/null 2>&1; then
      xdotool key ctrl+shift+v
    elif command -v wtype >/dev/null 2>&1; then
      wtype -M ctrl -M shift -k v
    fi
    ;;
esac
