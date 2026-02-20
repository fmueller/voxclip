# Voxclip

[![CI](https://github.com/fmueller/voxclip/actions/workflows/ci.yml/badge.svg)](https://github.com/fmueller/voxclip/actions/workflows/ci.yml)

Voxclip is a command-line tool for voice capture and transcription that runs entirely on your machine. It supports Linux and macOS. It uses open-source speech models so your audio never leaves your computer — no cloud APIs, no accounts, no network dependency. The default workflow records audio, transcribes it, and copies the result to your clipboard in a single command. I built it to voice-prompt coding agents: a fast, local way to speak instructions to AI coding tools without switching context.

## Installation

Voxclip is installed from release artifacts. The CLI itself does **not** perform installation.

### Bootstrap installer (`curl | sh`)

```bash
curl -fsSL https://raw.githubusercontent.com/fmueller/voxclip/main/scripts/install-voxclip.sh | sh
```

Optional: pin to a specific release tag:

```bash
curl -fsSL https://raw.githubusercontent.com/fmueller/voxclip/main/scripts/install-voxclip.sh | sh -s -- --version vX.Y.Z
```

The installer:

- detects OS/arch
- downloads the matching release archive
- verifies SHA256 via `checksums.txt`
- installs `voxclip` to `~/.local/bin` by default
- installs bundled whisper runtime to `~/.local/libexec/whisper/whisper-cli`

## Prerequisites

Voxclip requires a recording backend to capture audio.

**macOS:**

- `ffmpeg` — required for audio recording (`brew install ffmpeg`)

**Linux** (at least one of):

- `pw-record` (PipeWire) — preferred, usually pre-installed on modern distros
- `arecord` (ALSA utils) — fallback (`apt install alsa-utils` / `dnf install alsa-utils`)
- `ffmpeg` — last resort fallback (`apt install ffmpeg` / `dnf install ffmpeg`)

## Quickstart

```bash
voxclip setup
voxclip
```

- `voxclip setup` downloads and verifies the selected model.
- `voxclip` runs the default flow: record -> transcribe -> copy.

## Examples

- For voice-prompt workflows with terminal coding agents (interactive sessions, hotkey auto-paste, and one-shot invocation), see `examples/README.md`.

## Commands

- `voxclip` default flow: record -> transcribe -> copy
- `voxclip record` record audio to WAV
- `voxclip transcribe <audio-file>` transcribe existing audio file
- `voxclip devices` list recording devices and backend diagnostics
- `voxclip setup` download and verify model assets

There is intentionally no `voxclip install` command.

## Global Flags

- `--verbose` enable verbose logs
- `--json` JSON log output
- `--no-progress` disable progress rendering
- `--model <name|path>` model name or custom model file path
- `--model-dir <path>` override model storage directory
- `--language <auto|en|de|...>` language selection
- `--auto-download` auto-download missing models (default: true)
- `--backend <auto|pw-record|arecord|ffmpeg>` preferred recording backend; if it fails, Voxclip tries the remaining available backends
- `--input <selector>` input device selector for the chosen backend (for example `:1` or `:2` on macOS `ffmpeg`/`avfoundation`)
- `--copy-empty` copy blank transcripts to clipboard (default: false)
- `--silence-gate` detect near-silent WAV audio and skip transcription (default: true)
- `--silence-threshold-dbfs` silence gate threshold in dBFS (default: -65)
- `--duration <duration>` record duration, e.g. `10s`; 0 means interactive start/stop (default: 0)
- `--immediate` start recording immediately without waiting for Enter (default: false)

## Model Storage

- Linux: `$XDG_DATA_HOME/voxclip/models` or `~/.local/share/voxclip/models`
- macOS: `~/Library/Application Support/voxclip/models`

Override with `--model-dir`.

## Bundled Whisper Engine

Voxclip expects a bundled whisper runtime in release archives:

```text
voxclip
libexec/whisper/whisper-cli
```

At runtime Voxclip resolves the engine relative to the `voxclip` executable.

If the engine is missing, Voxclip returns an actionable error asking for reinstall from an official release.

Known limitations:

- GPU offload availability depends on how the bundled `whisper-cli` is compiled for each platform.
- Voxclip keeps runtime behavior compatible with whisper defaults, but portability-first bundles may run CPU-only on some systems.
- If you require GPU acceleration everywhere, ship whisper binaries compiled with the relevant backend support for your target OS/driver stack.

No speech handling:

- If whisper returns no speech (for example `[BLANK_AUDIO]`), Voxclip prints a hint to check mic mute/input device.
- Blank transcripts are not copied to clipboard by default; use `--copy-empty` to force copying.
- For WAV inputs, Voxclip applies a silence gate before transcription to avoid common near-silence hallucinations (for example "you").

## Recording Backends

Linux backend order:

1. `pw-record`
2. `arecord`
3. `ffmpeg`

macOS backend:

1. `ffmpeg` (`avfoundation`)

Use `voxclip devices` for diagnostics and `--backend` to force a backend.

If recording starts using the wrong microphone (for example when iPhone/headphones are connected on macOS), run `voxclip devices`, find the desired input index, and pass it via `--input` (for example `--input ":1"` or `--input ":2"`).

## Development

### Development prerequisites

**Both platforms:**

- Go 1.26+
- `git`

**Building whisper-cli** (optional, for local packaging):

- `cmake` (`brew install cmake` / `apt install cmake` / `dnf install cmake`)

**Optional convenience:**

- [go-task](https://taskfile.dev) (`brew install go-task` / `dnf install go-task`) — for `Taskfile.yml` wrappers
- Ubuntu/Debian via apt requires adding Task's repo first:

```bash
curl -1sLf 'https://dl.cloudsmith.io/public/task/task/setup.deb.sh' | sudo -E bash
sudo apt install task
```

Go toolchain is the source of truth; Taskfile provides convenience wrappers.

```bash
task tidy
task build
task test
task test:integration
```

For release process and packaging details, see `RELEASING.md`.
For user-facing release notes, see `CHANGELOG.md`.
