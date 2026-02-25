# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

Initial development history before the first tagged release.

### Added

- `--input` flag now selects devices on all Linux backends: `--target` for pw-record and `-D` for arecord.

### Changed

- `--input` flag help text updated with cross-platform examples (macOS, PipeWire, ALSA).
- Flags are now strictly command-scoped; unsupported flags on a command now return an unknown-flag error instead of being silently accepted from inherited global flags.

### Previously added

- Native Go CLI with `voxclip`, `record`, `transcribe`, `devices`, and `setup` commands.
- One-command default flow to record audio, transcribe locally, and copy the transcript to the clipboard.
- `voxclip setup` workflow for model download and verification.
- `voxclip devices` diagnostics for recording backends and audio devices.
- Cross-platform recording backend support (`pw-record`, `arecord`, `ffmpeg` on Linux; `ffmpeg` on macOS).
- Silence-gate handling for WAV inputs to skip near-silent audio and reduce hallucinated transcripts.
- Preflight transcription-readiness checks before recording starts.
- Release installer support with OS/arch detection, checksum verification, and optional version pinning.

### Changed

- Installation is centered on official release artifacts and the installer script.
- Moved the bootstrap installer script to `scripts/install-voxclip.sh` for repository consistency.
- Default flow now removes temporary recordings after completion.
- Input device and recording format flags are handled consistently across commands.
- Recording now falls back across implemented backends at runtime; errors are surfaced only after all backend attempts fail.
- PipeWire duration recording no longer depends on `pw-record --duration`; Voxclip now stops timed recordings itself for broader version compatibility.
- Release archives now include `THIRD_PARTY_NOTICES.md` to document dependency licenses for redistribution.

[unreleased]: https://github.com/fmueller/voxclip/commits/main
