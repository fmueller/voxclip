# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

Initial development history before the first tagged release.

### Added

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

[unreleased]: https://github.com/fmueller/voxclip/commits/main
