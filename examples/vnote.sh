#!/bin/sh
# vnote.sh — append a timestamped voice note to a text file.
#
# Usage:
#   sh examples/vnote.sh
#
# Configure:
#   VNOTE_FILE       — output file (default: ~/voice-notes.txt)
#   VNOTE_DURATION   — recording duration (default: 15s)

NOTES_FILE="${VNOTE_FILE:-$HOME/voice-notes.txt}"

text=$(voxclip --no-progress --duration "${VNOTE_DURATION:-15s}" 2>/dev/null)

[ -n "$text" ] && printf '%s  %s\n' "$(date '+%Y-%m-%d %H:%M')" "$text" >> "$NOTES_FILE"
