# Voice-Prompt Integration Examples

These scripts integrate voxclip with CLI coding agents and system hotkeys for hands-free voice prompting.

## Shell Functions

### `vprompt` — command substitution (fixed duration)

Source the helper and use it inside `$(...)`:

```bash
source examples/vprompt.sh

claude "$(vprompt)"
aider --message "$(vprompt)"
```

Recording lasts for a fixed duration (default 10s, configurable via `VPROMPT_DURATION`). No TTY is required, so it works inside command substitution where stdin is not a terminal.

```bash
export VPROMPT_DURATION=15s
claude "$(vprompt)"
```

To load automatically, add to your `~/.bashrc` or `~/.zshrc`:

```bash
source /path/to/voxclip/examples/vprompt.sh
```

### `vprompt_interactive` — interactive mode (Enter to stop)

Source the helper and call it from a real terminal:

```bash
source examples/vprompt-interactive.sh

vprompt_interactive
# Recording starts immediately — speak, then press Enter when done.
# Transcript is copied to your clipboard.
# Paste with Cmd+V (macOS) or Ctrl+Shift+V (terminal).
```

This is useful when you are already in an agent session (e.g. Claude Code) and want to record a voice note in a split pane or tab, then paste the result.

## Auto-Paste Hotkey

`vpaste.sh` records your voice for a fixed duration, transcribes it locally, and simulates a paste keystroke into whatever window was active.

### What happens in practice

1. You are typing in a Claude Code terminal session.
2. You press a global keyboard shortcut (e.g. `Ctrl+Shift+V`).
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

1. Copy the script and make it executable:
   ```bash
   cp examples/vpaste.sh ~/.local/bin/vpaste
   chmod +x ~/.local/bin/vpaste
   ```
2. Open **Settings > Keyboard > Custom Shortcuts**.
3. Add a new shortcut with command `~/.local/bin/vpaste` and assign your preferred key.

### Linux setup (KDE)

1. Copy the script as above.
2. Open **System Settings > Shortcuts > Custom Shortcuts**.
3. Add a new command trigger pointing to `~/.local/bin/vpaste` and assign your preferred key.

### Requirements

- **macOS:** nothing extra — `osascript` is built-in and voxclip copies to clipboard via `pbcopy`.
- **Linux X11:** `xdotool` for simulating the paste keystroke (`apt install xdotool` / `dnf install xdotool`).
- **Linux Wayland:** `wtype` for simulating the paste keystroke (`apt install wtype`).

### Configuration

Set `VPROMPT_DURATION` to change the recording length:

```bash
export VPROMPT_DURATION=12s
```
