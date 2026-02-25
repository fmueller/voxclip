# Voxclip

[![CI](https://github.com/fmueller/voxclip/actions/workflows/ci.yml/badge.svg)](https://github.com/fmueller/voxclip/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/fmueller/voxclip)](https://github.com/fmueller/voxclip/blob/main/go.mod)
[![License](https://img.shields.io/github/license/fmueller/voxclip)](https://github.com/fmueller/voxclip/blob/main/LICENSE)

Voxclip is a command-line tool for voice capture and transcription on Linux and macOS. It runs locally with open-source speech models, and the default flow records audio, transcribes it, and copies the result to your clipboard in one command.

## Table of Contents

- [Installation](#installation)
- [Prerequisites](#prerequisites)
- [Quickstart](#quickstart)
- [Commands](#commands)
- [Common Flags](#common-flags)
- [Recording Backends](#recording-backends)
- [Troubleshooting](#troubleshooting)
- [Advanced Runtime Details](#advanced-runtime-details)
- [Examples](#examples)
- [Development](#development)

## Installation

Voxclip is installed from release artifacts. There is intentionally no `voxclip install` command.

### Installer script (recommended)

**Default**

```bash
curl -fsSL https://raw.githubusercontent.com/fmueller/voxclip/main/scripts/install-voxclip.sh | sh
```

**Pinned**

```bash
curl -fsSL https://raw.githubusercontent.com/fmueller/voxclip/main/scripts/install-voxclip.sh | sh -s -- --version vX.Y.Z
```

**Review first**

```bash
curl -fsSL https://raw.githubusercontent.com/fmueller/voxclip/main/scripts/install-voxclip.sh -o install-voxclip.sh
less install-voxclip.sh
sh install-voxclip.sh
```

The installer:

- detects your OS/arch and downloads the matching release archive
- verifies archive integrity with `checksums.txt`
- installs `voxclip` to `~/.local/bin` and whisper runtime to `~/.local/libexec/whisper/whisper-cli`

### Manual install from release archive

1. Download the matching archive for your OS/arch from [GitHub Releases](https://github.com/fmueller/voxclip/releases).
2. Extract it and keep this layout intact:

```text
voxclip
libexec/whisper/whisper-cli
```

3. Put `voxclip` on your `PATH` and preserve its relative path to `libexec/whisper/whisper-cli`.

## Prerequisites

Voxclip requires a recording backend to capture audio.

**macOS:**

- `ffmpeg` (required) - install with `brew install ffmpeg`

**Linux** (at least one):

- `pw-record` (PipeWire) - preferred, usually pre-installed on modern distros
- `arecord` (ALSA utils) - fallback (`apt install alsa-utils` / `dnf install alsa-utils`)
- `ffmpeg` - last resort fallback (`apt install ffmpeg` / `dnf install ffmpeg`)

## Quickstart

```bash
voxclip --version
voxclip setup
voxclip
```

- `voxclip setup` downloads and verifies the selected model.
- `voxclip` runs the default flow: record -> transcribe -> copy.
- Expected result: transcript prints in the terminal and is copied to your clipboard.

## Commands

- `voxclip` run the default flow (record -> transcribe -> copy)
- `voxclip record` record audio to WAV
- `voxclip transcribe <audio-file>` transcribe existing audio
- `voxclip devices` list recording devices and backend diagnostics
- `voxclip setup` download and verify model assets

For complete command and flag reference, run `voxclip --help` and `voxclip <command> --help`.

## Common Flags

- `--model <name|path>` select a model name or local model path
- `--language <auto|en|de|...>` set transcription language
- `--backend <auto|pw-record|arecord|ffmpeg>` choose recording backend
- `--input <selector>` choose input device (for example `:1` on macOS, a PipeWire node ID for `pw-record`, or `hw:1,0` for `arecord`)
- `--duration <duration>` set fixed recording duration, e.g. `10s`
- `--immediate` start recording immediately
- `--verbose` enable verbose logs

For all global flags, run `voxclip --help`.

## Recording Backends

Linux backend order:

1. `pw-record`
2. `arecord`
3. `ffmpeg`

macOS backend:

1. `ffmpeg` (`avfoundation`)

Use `voxclip devices` for diagnostics and `--backend` to force a backend.

If recording starts from the wrong microphone, run `voxclip devices`, find the desired input identifier, and pass it with `--input`:

- **macOS (ffmpeg/avfoundation):** `--input ":1"` or `--input ":2"` (device index)
- **Linux (pw-record):** `--input "42"` (PipeWire node ID from `pw-cli ls Node`)
- **Linux (arecord):** `--input "hw:1,0"` (ALSA PCM device from `arecord -L`)

## Troubleshooting

- No speech detected (`[BLANK_AUDIO]`) -> check mute state, input device, and microphone gain.
- Blank transcript not copied -> use `--copy-empty`.
- Wrong microphone selected -> run `voxclip devices` and set `--input`.
- Near-silent WAV false positives -> debug with `--silence-gate=false`, then tune `--silence-threshold-dbfs`.
- Missing recording backend -> install one of `pw-record`, `arecord`, or `ffmpeg`.
- Missing whisper runtime -> reinstall an official release so `libexec/whisper/whisper-cli` is present.

## Advanced Runtime Details

Voxclip expects this bundled layout in release archives:

```text
voxclip
libexec/whisper/whisper-cli
```

- Runtime lookup: Voxclip resolves `libexec/whisper/whisper-cli` relative to `voxclip`.
- Model storage (Linux): `$XDG_DATA_HOME/voxclip/models` or `~/.local/share/voxclip/models`
- Model storage (macOS): `~/Library/Application Support/voxclip/models`
- Model storage override: `--model-dir`
- GPU offload depends on how bundled `whisper-cli` is built per platform.
- Portability-first bundles may run CPU-only on some systems.
- If you require GPU acceleration everywhere, ship whisper binaries compiled for your target backend and driver stack.

## Examples

For voice-prompt workflows with terminal coding agents (interactive sessions, hotkey auto-paste, and one-shot invocation), see `examples/README.md`.

## Development

For contributors working on Voxclip source and release tooling.

### Development prerequisites

**Both platforms:**

- Go 1.26+
- `git`

**Building whisper-cli** (optional, for local packaging):

- `cmake` (`brew install cmake` / `apt install cmake` / `dnf install cmake`)

**Optional convenience:**

- [go-task](https://taskfile.dev) (`brew install go-task` / `dnf install go-task`) for `Taskfile.yml` wrappers
- Ubuntu/Debian via apt requires adding Task's repo first:

```bash
curl -1sLf 'https://dl.cloudsmith.io/public/task/task/setup.deb.sh' | sudo -E bash
sudo apt install task
```

After cloning, stage a host-local whisper runtime once:

```bash
task whisper:build
```

Without Task installed:

```bash
./scripts/build-whisper-cli.sh
```

Go toolchain is the source of truth; Taskfile provides convenience wrappers:

```bash
task tidy
task build
task test
task test:integration
```

For release process and packaging details, see `RELEASING.md`.
For user-facing release notes, see `CHANGELOG.md`.
