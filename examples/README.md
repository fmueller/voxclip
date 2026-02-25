# Voice-Prompt Integration Examples

These scripts integrate voxclip with terminal coding agents and system hotkeys for voice prompting.

They are agent-agnostic and work with tools like `claude`, `codex`, `opencode`, `crush`, and `aider`.

## Interactive sessions (agent already open)

Interactive means your coding agent is already running in a terminal session and you use voxclip to insert text into that live prompt instead of typing everything.

### `vprompt_interactive` — interactive mode (Enter to stop)

Source the helper and call it from a real terminal:

```bash
source examples/vprompt-interactive.sh

vprompt_interactive
# Recording starts immediately — speak, then press Enter when done.
# Transcript is copied to your clipboard.
# Paste with Cmd+V (macOS) or Ctrl+Shift+V (terminal).
```

This is useful when you are already in an agent session and want to record a voice note in a split pane or tab, then paste the result.

To load automatically, add to your `~/.bashrc` or `~/.zshrc`:

```bash
source /path/to/voxclip/examples/vprompt-interactive.sh
```

Example mixed text + voice flow in one live session:

1. Type context in the agent prompt.
2. Capture a voice chunk with `vprompt_interactive` in another pane/tab.
3. Paste transcript into the agent prompt.
4. Type final constraints and submit.

## Auto-Paste Hotkey

`vpaste.sh` records your voice for a fixed duration, transcribes it locally, and simulates a paste keystroke into whatever window was active.

### What happens in practice

1. You are typing in a terminal coding agent session.
2. You press a global keyboard shortcut (e.g. `Ctrl+Shift+V` on macOS, `Super+V` on Linux).
3. The system runs `vpaste.sh` in the background — your microphone activates and records for N seconds (default 8s).
4. Voxclip transcribes the audio locally using the bundled whisper engine.
5. The transcript is copied to your clipboard.
6. The script simulates a paste keystroke (`Cmd+V` on macOS, `Ctrl+Shift+V` on Linux).
7. The transcript appears in your terminal at the cursor, as if you typed it.

The recording happens silently — no terminal window opens, no prompts appear. You just speak for N seconds after pressing the hotkey. The trade-off is a fixed recording duration (you cannot stop early), but this is what makes the fully hands-free workflow possible.

### macOS setup (Shortcuts app)

1. Open **Shortcuts.app** and create a new shortcut named "Voice Prompt".
2. Add the action **Run Shell Script**, set shell to `/bin/zsh`, and paste the contents of `vpaste.sh`.
3. Go to **System Settings > Keyboard > Keyboard Shortcuts > Services** (or App Shortcuts) and assign a key combination (e.g. `Ctrl+Shift+V`) to the "Voice Prompt" shortcut.

### Linux setup (GNOME)

> **Note:** On Linux, the hotkey must not be `Ctrl+Shift+V` because the script uses that same combo to simulate pasting into the terminal. Use a different shortcut such as `Super+V` or `Ctrl+Alt+V`.

1. Copy the script and make it executable:
   ```bash
   cp examples/vpaste.sh ~/.local/bin/vpaste
   chmod +x ~/.local/bin/vpaste
   ```
2. Open **Settings > Keyboard > Custom Shortcuts**.
3. Add a new shortcut with command `~/.local/bin/vpaste` and assign your preferred key (e.g. `Super+V`).

### Linux setup (KDE)

1. Copy the script as above.
2. Open **System Settings > Shortcuts > Custom Shortcuts**.
3. Add a new command trigger pointing to `~/.local/bin/vpaste` and assign your preferred key (e.g. `Super+V`).

### Requirements

- **macOS:** nothing extra — `osascript` is built-in and voxclip copies to clipboard via `pbcopy`.
- **Linux X11:** `xdotool` for simulating the paste keystroke (`apt install xdotool` / `dnf install xdotool`).
- **Linux Wayland:** `wtype` for simulating the paste keystroke (`apt install wtype`).

### Configuration

Set `VPROMPT_DURATION` to change the recording length:

```bash
export VPROMPT_DURATION=12s
```

For faster transcription in voice-prompt workflows, use `--model tiny` (~40 MB download instead of ~465 MB for the default `small` model). Edit the script or override the model flag:

```bash
voxclip --model tiny --no-progress --duration 8s 2>/dev/null
```

For non-English prompts, set the language explicitly — auto-detection can be unreliable for short utterances:

```bash
voxclip --language de --no-progress --duration 8s 2>/dev/null
```

## One-shot invocation (non-interactive)

### `vprompt` — command substitution (fixed duration)

Source the helper and use it inside `$(...)`:

```bash
source examples/vprompt.sh

your-agent-cli "$(vprompt)"
```

Recording lasts for a fixed duration (default 10s, configurable via `VPROMPT_DURATION`). No TTY is required, so it works inside command substitution where stdin is not a terminal.

For example:

```bash
export VPROMPT_DURATION=15s
claude "$(vprompt)"
aider --message "$(vprompt)"
```

To load automatically, add to your `~/.bashrc` or `~/.zshrc`:

```bash
source /path/to/voxclip/examples/vprompt.sh
```

## Voice notes

### `vnote.sh` — append timestamped voice notes to a file

A simple script that records, transcribes, and appends the result to a text file with a timestamp:

```bash
sh examples/vnote.sh
```

Each note is appended as a single line:

```
2026-02-25 14:30  Refactor the download package to use context timeouts.
2026-02-25 15:12  Check if the silence gate threshold needs adjusting for USB mics.
```

Configure via environment variables:

```bash
export VNOTE_FILE=~/project-notes.txt   # default: ~/voice-notes.txt
export VNOTE_DURATION=20s               # default: 15s
```
