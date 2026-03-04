---
title: Getting Started
weight: 2
---

After [installing Voxclip](../installation), follow these steps to make your first transcription.

{{% steps %}}

### Verify the installation

```bash
voxclip version
```

### Download a speech model

```bash
voxclip setup
```

This downloads and verifies the default model. Models are stored in:

- **Linux:** `$XDG_DATA_HOME/voxclip/models` or `~/.local/share/voxclip/models`
- **macOS:** `~/Library/Application Support/voxclip/models`

### Record and transcribe

```bash
voxclip
```

This runs the default flow:

1. **Record** — captures audio from your microphone
2. **Transcribe** — processes the audio with the speech model
3. **Copy** — puts the transcript on your clipboard

Press `Ctrl+C` to stop recording. The transcript prints in the terminal and is copied to your clipboard.

{{% /steps %}}

## Next steps

- Explore the full [command reference](../commands)
- Configure your [recording backend](../recording-backends)
- Check [troubleshooting](../troubleshooting) if something goes wrong
