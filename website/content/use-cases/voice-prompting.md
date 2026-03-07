---
title: Voice Prompting for Coding Agents
weight: 1
---

You are deep in a coding session with Claude Code, aider, or another terminal agent. You know exactly what you want — a refactor across three files, a test for an edge case, a design question — but typing a detailed prompt takes longer than the thought itself. With Voxclip you speak the instruction, and the transcript lands on your clipboard ready to paste.

## How it works

1. You speak into your microphone.
2. Voxclip records and transcribes locally using an open-source speech model.
3. The transcript is copied to your clipboard.
4. You paste it into your coding agent's prompt.

No audio leaves your machine. No cloud API. Works offline.

## Quick start

Source the helper function and start prompting by voice:

```bash
source examples/vprompt-interactive.sh

vprompt_interactive
# Speak, then press Enter to stop.
# Paste into your agent with Cmd+V / Ctrl+Shift+V.
```

Add it to your shell profile (`~/.bashrc` or `~/.zshrc`) so it is always available:

```bash
source /path/to/voxclip/examples/vprompt-interactive.sh
```

## Variants

**One-shot invocation** — pass a voice prompt directly as a command argument:

```bash
source examples/vprompt.sh

claude "$(vprompt)"
aider --message "$(vprompt)"
```

Recording lasts a fixed duration (default 10 seconds, configurable via `VPROMPT_DURATION`).

**Auto-paste hotkey** — bind `vpaste.sh` to a global shortcut so you can speak and have the transcript pasted into the active window automatically, no terminal interaction needed.

## Next steps

See the [Examples reference](/docs/examples#coding-agents) for platform-specific hotkey setup (macOS/GNOME/KDE), all environment variables, and the mixed text-plus-voice workflow.
