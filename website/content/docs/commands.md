---
title: Commands & Flags
weight: 3
---

## Commands

| Command | Description |
|---------|-------------|
| `voxclip` | Run the default flow (record → transcribe → copy) |
| `voxclip record` | Record audio to WAV |
| `voxclip transcribe <audio-file>` | Transcribe existing audio |
| `voxclip devices` | List recording devices and backend diagnostics |
| `voxclip setup` | Download and verify model assets |
| `voxclip version` | Show version information |

For complete flag reference, run `voxclip --help` and `voxclip <command> --help`.

## Default-flow flags

These flags apply to the main `voxclip` command:

| Flag | Description |
|------|-------------|
| `--model <name\|path>` | Select a model name (tiny, base, small, medium, large-v3) or local model path (default: small on macOS, tiny on Linux) |
| `--model-dir <path>` | Override model storage directory |
| `--language <auto\|en\|de\|...>` | Set transcription language |
| `--auto-download` | Automatically download a missing model |
| `--backend <auto\|pw-record\|arecord\|ffmpeg>` | Choose recording backend |
| `--input <selector>` | Choose input device |
| `--input-format <pulse\|alsa>` | Force ffmpeg input format on Linux |
| `--copy-empty` | Copy blank transcripts to clipboard |
| `--copy-newline` | Append a trailing newline to the clipboard text |
| `--silence-gate` | Enable near-silent WAV detection before transcription |
| `--silence-threshold-dbfs <value>` | Set silence-gate threshold |
| `--duration <duration>` | Set fixed recording duration (e.g. `10s`) |
| `--immediate` | Start recording immediately |
| `--pid-file <path>` | Write PID to file and wait for SIGUSR1 to stop recording (for toggle-style hotkey workflows) |
| `--no-progress` | Disable spinner/progress indicators |
| `--verbose` | Enable verbose logs |
| `--json` | Output logs in JSON format |

## Command-specific flags

Each subcommand has its own flags:

- **`voxclip record --help`** — recording-only flags such as `--output`
- **`voxclip transcribe --help`** — transcription/copy flags such as `--copy`
- **`voxclip setup --help`** — model setup flags only
- **`voxclip devices --help`** — no operational flags

## Input device selection

{{< tabs items="macOS,Linux (PipeWire),Linux (ALSA)" >}}

{{< tab >}}
```bash
voxclip --input ":1"
```
Use the device index from `voxclip devices`.
{{< /tab >}}

{{< tab >}}
```bash
voxclip --input "42"
```
Use the PipeWire node ID from `pw-cli ls Node`.
{{< /tab >}}

{{< tab >}}
```bash
voxclip --input "hw:1,0"
```
Use the ALSA PCM device from `arecord -L`.
{{< /tab >}}

{{< /tabs >}}
