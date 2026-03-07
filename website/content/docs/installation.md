---
title: Installation
weight: 1
---

Voxclip is installed from release artifacts. There is intentionally no `voxclip install` command.

## Installer script (recommended)

**Default install:**

```bash
curl -fsSL https://raw.githubusercontent.com/fmueller/voxclip/main/scripts/install-voxclip.sh | sh
```

**Pinned version:**

```bash
curl -fsSL https://raw.githubusercontent.com/fmueller/voxclip/main/scripts/install-voxclip.sh | sh -s -- --version vX.Y.Z
```

**Review the script first:**

```bash
curl -fsSL https://raw.githubusercontent.com/fmueller/voxclip/main/scripts/install-voxclip.sh -o install-voxclip.sh
less install-voxclip.sh
sh install-voxclip.sh
```

The installer:

- Detects your OS/arch and downloads the matching release archive
- Verifies archive integrity with `checksums.txt`
- Installs `voxclip` to `~/.local/bin` and the whisper runtime to `~/.local/libexec/whisper/whisper-cli`

Installer downloads come from **GitHub Releases assets** for the selected tag. The target machine does not build `whisper.cpp` — the release archive already contains a prebuilt `whisper-cli` binary.

## Manual install from release archive

1. Download the matching archive for your OS/arch from [GitHub Releases](https://github.com/fmueller/voxclip/releases).
2. Extract it and keep this layout intact:

```text
voxclip
libexec/whisper/whisper-cli
```

3. Put `voxclip` on your `PATH` and preserve its relative path to `libexec/whisper/whisper-cli`.

## Prerequisites

Voxclip requires a recording backend to capture audio.

{{< tabs items="macOS,Linux" >}}

{{< tab >}}
`ffmpeg` is required — install with:

```bash
brew install ffmpeg
```
{{< /tab >}}

{{< tab >}}
At least one of the following (in preference order):

- **`pw-record`** (PipeWire) — preferred, usually pre-installed on modern distros
- **`arecord`** (ALSA utils) — fallback (`apt install alsa-utils` / `dnf install alsa-utils`)
- **`ffmpeg`** — last resort (`apt install ffmpeg` / `dnf install ffmpeg`)
{{< /tab >}}

{{< /tabs >}}
