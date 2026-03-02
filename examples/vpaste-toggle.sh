#!/bin/sh
# vpaste-toggle.sh — toggle-style voice-to-paste: press hotkey to start, press again to stop.
#
# First invocation: starts voxclip recording, blocks until SIGUSR1 is received.
# Second invocation: sends SIGUSR1 to the running instance and exits.
# After recording stops, the first instance transcribes, copies to clipboard,
# and the script simulates a paste keystroke.
#
# Usage:
#   Bind this script to a global hotkey (see examples/README.md).
#   Press the hotkey to start recording, press again to stop and paste.
#
# Configure model (default: small):
#   export VOXCLIP_MODEL=tiny

set -e

export PATH="$HOME/.local/bin:/opt/homebrew/bin:/usr/local/bin:$PATH"

PID_FILE="${XDG_RUNTIME_DIR:-/tmp}/voxclip-toggle.pid"

# If a recording is already running, stop it and exit.
if [ -f "$PID_FILE" ] && kill -0 "$(cat "$PID_FILE")" 2>/dev/null; then
  kill -USR1 "$(cat "$PID_FILE")"
  exit 0
fi

# Start recording; blocks until SIGUSR1 stops it.
voxclip --pid-file "$PID_FILE" --model "${VOXCLIP_MODEL:-small}" --no-progress 2>/dev/null || {
  echo "vpaste-toggle: recording failed; is voxclip installed?" >&2
  exit 1
}

# Simulate paste keystroke.
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
