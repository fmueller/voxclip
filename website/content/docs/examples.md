---
title: Examples
weight: 6
---

{{< callout type="info" >}}
Looking for a quick overview? Start with the [Use Cases](/use-cases) section for narrative walkthroughs, then come back here for detailed setup and configuration.
{{< /callout >}}

Voxclip ships with ready-made scripts for voice-prompt workflows with terminal coding agents and system-wide voice-to-text. All processing runs locally — no cloud APIs, no data leaves your machine.

The full script sources live in the [`examples/`](https://github.com/fmueller/voxclip/tree/main/examples) directory.

## Coding Agents

Use your voice to prompt terminal coding agents like `claude`, `aider`, `codex`, `opencode`, or `crush`.

### Interactive mode (`vprompt_interactive`)

Record in a split pane or tab, then paste the transcript into your agent session.

```bash
source examples/vprompt-interactive.sh

vprompt_interactive
# Recording starts immediately — speak, then press Enter when done.
# Transcript is copied to your clipboard.
# Paste with Cmd+V (macOS) or Ctrl+Shift+V (terminal).
```

To load automatically, add to your `~/.bashrc` or `~/.zshrc`:

```bash
source /path/to/voxclip/examples/vprompt-interactive.sh
```

**Example mixed text + voice flow:**

1. Type context in the agent prompt.
2. Capture a voice chunk with `vprompt_interactive` in another pane/tab.
3. Paste transcript into the agent prompt.
4. Type final constraints and submit.

### One-shot invocation (`vprompt`)

Use command substitution to pass a voice prompt directly:

```bash
source examples/vprompt.sh

claude "$(vprompt)"
aider --message "$(vprompt)"
```

Recording lasts for a fixed duration (default 10s, configurable via `VPROMPT_DURATION`). No TTY is required, so it works inside `$(...)` where stdin is not a terminal.

```bash
export VPROMPT_DURATION=15s
```

### Auto-paste hotkey (`vpaste.sh`)

Record your voice for a fixed duration, transcribe locally, and simulate a paste keystroke into whatever window was active. Fully hands-free — no terminal window opens, no prompts appear.

**What happens:**

1. Press a global keyboard shortcut.
2. Your microphone records for N seconds (default 8s).
3. Voxclip transcribes the audio locally.
4. The transcript is copied to your clipboard and pasted automatically.

{{< tabs items="macOS,Linux (GNOME),Linux (KDE)" >}}

{{< tab >}}
1. Open **Shortcuts.app** and create a new shortcut named "Voice Prompt".
2. Add the action **Run Shell Script**, set shell to `/bin/zsh`, and paste the contents of `vpaste.sh`.
3. Go to **System Settings > Keyboard > Keyboard Shortcuts > Services** and assign a key combination (e.g. `Ctrl+Shift+V`).
{{< /tab >}}

{{< tab >}}
> **Note:** On Linux, the hotkey must not be `Ctrl+Shift+V` because the script uses that same combo to simulate pasting. Use a different shortcut such as `Super+V`.

1. Copy the script:
   ```bash
   cp examples/vpaste.sh ~/.local/bin/vpaste
   chmod +x ~/.local/bin/vpaste
   ```
2. Open **Settings > Keyboard > Custom Shortcuts**.
3. Add a new shortcut with command `~/.local/bin/vpaste` and assign your preferred key (e.g. `Super+V`).
{{< /tab >}}

{{< tab >}}
1. Copy the script:
   ```bash
   cp examples/vpaste.sh ~/.local/bin/vpaste
   chmod +x ~/.local/bin/vpaste
   ```
2. Open **System Settings > Shortcuts > Custom Shortcuts**.
3. Add a new command trigger pointing to `~/.local/bin/vpaste` and assign your preferred key (e.g. `Super+V`).
{{< /tab >}}

{{< /tabs >}}

**Configuration:**

```bash
export VPROMPT_DURATION=12s
```

**Requirements:**
- **macOS:** nothing extra — `osascript` is built-in and voxclip copies to clipboard via `pbcopy`.
- **Linux X11:** `xclip` for clipboard writes and `xdotool` for simulating the paste keystroke (`apt install xclip xdotool` / `dnf install xclip xdotool`).
- **Linux Wayland:** `wl-copy` for clipboard writes and `wtype` for simulating the paste keystroke (`apt install wl-clipboard wtype`).

Why both steps are needed: these hotkey scripts only simulate the paste keypress; they do not place text on the clipboard themselves. `voxclip` performs the copy operation, then the script triggers paste into the active window.

## General Use Cases

### System-wide voice-to-text shortcut

Use `vpaste.sh` or `vpaste-toggle.sh` as a global hotkey to dictate into any application — browser, email, chat, notes — without cloud APIs. All processing runs locally on your machine.

### Toggle-style hotkey (`vpaste-toggle.sh`)

Press the hotkey once to start recording, press it again to stop. No fixed duration needed.

**How it works:**

1. **First press:** starts `voxclip --pid-file ...` which begins recording and blocks.
2. **Second press:** detects the running instance via the PID file, sends `SIGUSR1`, and exits immediately.
3. The first instance stops recording, transcribes, copies to clipboard, and simulates a paste keystroke.

{{< tabs items="macOS,Linux" >}}

{{< tab >}}
1. Open **Shortcuts.app** and create a new shortcut.
2. Add the action **Run Shell Script**, set shell to `/bin/zsh`, and paste the contents of `vpaste-toggle.sh`.
3. Assign a keyboard shortcut in **System Settings > Keyboard > Keyboard Shortcuts > Services**.
{{< /tab >}}

{{< tab >}}
```bash
cp examples/vpaste-toggle.sh ~/.local/bin/vpaste-toggle
chmod +x ~/.local/bin/vpaste-toggle
```
Then bind `~/.local/bin/vpaste-toggle` to a global hotkey:
- **GNOME:** Settings > Keyboard > Custom Shortcuts
- **KDE:** System Settings > Shortcuts > Custom Shortcuts
{{< /tab >}}

{{< /tabs >}}

**Configuration:**

```bash
export VOXCLIP_MODEL=tiny
```

**Requirements:** same as `vpaste.sh` above.

### Voice notes (`vnote.sh`)

Append timestamped transcriptions to a file:

```bash
sh examples/vnote.sh
```

Each note is appended as a single line:

```
2026-02-25 14:30  Refactor the download package to use context timeouts.
2026-02-25 15:12  Check if the silence gate threshold needs adjusting for USB mics.
```

**Configuration:**

```bash
export VNOTE_FILE=~/project-notes.txt   # default: ~/voice-notes.txt
export VNOTE_DURATION=20s               # default: 15s
```
