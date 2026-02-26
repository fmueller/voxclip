# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-02-26

### Added

- Native Go CLI with `voxclip`, `record`, `transcribe`, `devices`, and `setup` commands.
- One-command default flow to record audio, transcribe locally, and copy the transcript to the clipboard.
- `voxclip setup` workflow for model download and verification.
- `voxclip devices` diagnostics for recording backends and audio devices.
- Cross-platform recording backend support (`pw-record`, `arecord`, `ffmpeg` on Linux; `ffmpeg` on macOS).
- `--input` flag selects the recording device across all supported backends.
- Automatic silence detection to skip near-silent audio and reduce hallucinated transcripts.
- Preflight transcription-readiness checks before recording starts.
- Release installer support with OS/arch detection, checksum verification, and optional version pinning.

### Changed

- `--input` flag help text updated with cross-platform examples (macOS, PipeWire, ALSA).
- Flags are now strictly command-scoped; unsupported flags on a command now return an unknown-flag error instead of being silently accepted from inherited global flags.
- Installation is centered on official release artifacts and the installer script.
- Default flow now removes temporary recordings after completion.
- Input device and recording format flags are handled consistently across commands.
- Recording now falls back across implemented backends at runtime; errors are surfaced only after all backend attempts fail.
- PipeWire duration recording no longer depends on `pw-record --duration`; voxclip now stops timed recordings itself for broader version compatibility.
- Release archives now include `THIRD_PARTY_NOTICES.md` to document dependency licenses for redistribution.

[1.0.0]: https://github.com/fmueller/voxclip/releases/tag/v1.0.0
