---
title: Troubleshooting
weight: 5
---

## Common issues

{{< callout type="info" >}}
Run `voxclip devices` first to check backend availability and input device detection.
{{< /callout >}}

### No speech detected (`[BLANK_AUDIO]`)

Check your mute state, input device selection, and microphone gain. Make sure the correct microphone is active and not muted at the OS level.

### Blank transcript not copied to clipboard

By default, blank transcripts are not copied. Use the `--copy-empty` flag if you want blank results on the clipboard:

```bash
voxclip --language en --copy-empty
```

### Wrong microphone selected

Run `voxclip devices` to list available inputs, then specify the correct one with `--input`:

```bash
# macOS
voxclip --input ":1" --language en

# Linux (PipeWire)
voxclip --input "42" --language en

# Linux (ALSA)
voxclip --input "hw:1,0" --language en
```

### Near-silent WAV false positives

If the silence gate is triggering incorrectly, debug by disabling it first:

```bash
voxclip --language en --silence-gate=false
```

Then tune the threshold:

```bash
voxclip --language en --silence-threshold-dbfs -35
```

### Missing recording backend

Install one of the supported backends:

{{< tabs items="macOS,Linux" >}}

{{< tab >}}
```bash
brew install ffmpeg
```
{{< /tab >}}

{{< tab >}}
```bash
# PipeWire (preferred, usually pre-installed)
# ALSA
apt install alsa-utils    # or dnf install alsa-utils
# ffmpeg
apt install ffmpeg        # or dnf install ffmpeg
```
{{< /tab >}}

{{< /tabs >}}

### Clipboard not working on Linux

Clipboard copy on Linux requires either `wl-copy` (Wayland sessions) or `xclip` (X11/XWayland sessions):

```bash
# Wayland
apt install wl-clipboard   # provides wl-copy

# X11 / XWayland
apt install xclip
```

### Transcript prints to terminal but isn't on clipboard

Transcript output to stdout is intentional — it gives you immediate visibility and allows piping into other commands. Clipboard copy is an additional convenience, not a replacement. If the transcript appears in your terminal but isn't on the clipboard, check the clipboard tool requirements above.

### Missing whisper runtime

Reinstall from an official release so that `libexec/whisper/whisper-cli` is present alongside the `voxclip` binary. See the [installation guide](../installation) for details.

To override the whisper runtime location, set the `VOXCLIP_WHISPER_PATH` environment variable:

```bash
export VOXCLIP_WHISPER_PATH=/path/to/whisper-cli
```
