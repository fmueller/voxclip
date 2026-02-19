# Voxclip

Voxclip is a Go CLI for local voice capture and transcription.

- CLI framework: `github.com/spf13/cobra`
- Logging: `go.uber.org/zap`
- Progress bars: `github.com/schollz/progressbar/v3`
- Tests: standard `testing` + `github.com/stretchr/testify/require`

## Installation

Voxclip is installed from release artifacts. The CLI itself does **not** perform installation.

### Bootstrap installer (`curl | sh`)

```bash
curl -fsSL https://raw.githubusercontent.com/fmueller/voxclip/main/install-voxclip.sh | sh
```

Optional version pin:

```bash
curl -fsSL https://raw.githubusercontent.com/fmueller/voxclip/main/install-voxclip.sh | sh -s -- --version v1.2.3
```

The installer:

- detects OS/arch
- downloads the matching release archive
- verifies SHA256 via `checksums.txt`
- installs `voxclip` to `~/.local/bin` by default
- installs bundled whisper runtime to `~/.local/libexec/whisper/whisper-cli`

### Homebrew (planned via tap)

Releases are configured for Homebrew publishing through GoReleaser (`.goreleaser.yml`, `brews` section).

Target tap repo layout:

- `homebrew-tap/`
  - `Formula/voxclip.rb`

## Quickstart

```bash
voxclip setup
voxclip
```

- `voxclip setup` downloads and verifies the selected model.
- `voxclip` runs the default flow: record -> transcribe -> copy.

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
- `--backend <auto|pw-record|arecord|ffmpeg>` recording backend override
- `--copy-empty` copy blank transcripts to clipboard (default: false)
- `--silence-gate` detect near-silent WAV audio and skip transcription (default: true)
- `--silence-threshold-dbfs` silence gate threshold in dBFS (default: -65)

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

## Development

Go toolchain is the source of truth; Taskfile provides convenience wrappers.

```bash
task tidy
task build
task test
task test:integration
```

Release-related tasks:

```bash
task whisper:build
task release:check
task release:dry
```

## GitHub Actions

Two workflows are included:

- `.github/workflows/ci.yml`
  - runs on `push` to `main` and on PRs
  - checks gofmt, runs `go vet`, unit tests, integration-tag tests, CLI help smoke tests, and installer script validation
- `.github/workflows/release.yml`
  - runs on tags (`v*`) and manual dispatch
  - builds `whisper-cli` for `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`
  - stages binaries into `packaging/whisper/<os>_<arch>/whisper-cli`
  - runs GoReleaser to publish release artifacts and checksums

Required repository secret for Homebrew publishing:

- `HOMEBREW_TAP_GITHUB_TOKEN`

## Release Packaging Notes

GoReleaser expects platform-specific whisper binaries before packaging:

- `packaging/whisper/linux_amd64/whisper-cli`
- `packaging/whisper/linux_arm64/whisper-cli`
- `packaging/whisper/darwin_amd64/whisper-cli`
- `packaging/whisper/darwin_arm64/whisper-cli`

These files are included in release archives under `libexec/whisper/whisper-cli`.

You can stage your current host platform binary locally with:

```bash
./scripts/build-whisper-cli.sh
```

For tagged releases, GitHub Actions builds and stages all four target binaries automatically.
