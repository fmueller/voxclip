---
title: System-Wide Voice-to-Text
weight: 2
---

You want to dictate into any application — a browser form, an email draft, a chat window — without switching to a dedicated dictation app or sending audio to the cloud. With Voxclip and a global hotkey, you press a key combination, speak, and the transcript is pasted into whatever window is active.

## How it works

1. Press your global keyboard shortcut.
2. Your microphone records for a set duration (or until you press the hotkey again).
3. Voxclip transcribes the audio locally.
4. The transcript is copied to your clipboard and pasted automatically.

Everything runs on your machine. No account, no network request.

## Quick start

Copy the script and bind it to a hotkey:

```bash
cp examples/vpaste.sh ~/.local/bin/vpaste
chmod +x ~/.local/bin/vpaste
```

Then assign `~/.local/bin/vpaste` to a global keyboard shortcut in your desktop settings (e.g. `Super+V`).

## Variants

**Fixed-duration** (`vpaste.sh`) — records for a set number of seconds (default 8, configurable via `VPROMPT_DURATION`), then transcribes and pastes.

**Toggle-style** (`vpaste-toggle.sh`) — press the hotkey once to start recording, press it again to stop. No fixed duration needed. Useful when you do not know in advance how long you will speak.

## Next steps

See the [Examples reference](/docs/examples#general-use-cases) for platform-specific hotkey setup on macOS, GNOME, and KDE, configuration options, and system requirements.
