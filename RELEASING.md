# Releasing Voxclip

This document is for maintainers and release engineers.

## Source-of-truth files

- Release workflow: `.github/workflows/release.yml`
- CI validation workflow: `.github/workflows/ci.yml`
- GoReleaser config: `.goreleaser.yml`
- Whisper staging details: `packaging/whisper/README.md`
- User-facing CLI docs: `README.md`

## Release-related commands

Use these commands to validate release setup locally:

```bash
task whisper:build
task release:check
task release:dry
```

Equivalent GoReleaser commands:

```bash
goreleaser check
goreleaser release --snapshot --clean
```

## Bundled whisper binaries

Before packaging, platform-specific whisper binaries must be staged under `packaging/whisper/...`.

For exact expected paths and host-local staging instructions, see `packaging/whisper/README.md`.

Release archives include the staged binary as:

```text
libexec/whisper/whisper-cli
```

## GitHub Actions release flow

`.github/workflows/release.yml` runs on version tags (`v*`) and manual dispatch.

The workflow builds and stages `whisper-cli` for:

- `linux/amd64`
- `linux/arm64`
- `darwin/amd64`
- `darwin/arm64`

It then runs GoReleaser to publish release archives and `checksums.txt`.
