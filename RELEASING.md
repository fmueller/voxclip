# Releasing Voxclip

This document is for maintainers and release engineers.

## Source-of-truth files

- Release workflow: `.github/workflows/release.yml`
- CI validation workflow: `.github/workflows/ci.yml`
- GoReleaser config: `.goreleaser.yml`
- Changelog and release notes: `CHANGELOG.md`
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

## Changelog policy

`CHANGELOG.md` is user-facing release notes.

- Keep entries focused on CLI behavior, flags, workflows, installation UX, and compatibility notes.
- Do not list internal-only refactors, CI plumbing, or dependency bumps unless users are affected.
- Keep upcoming changes under `## [Unreleased]` and move them into a dated version section when tagging a release.

## Bundled whisper binaries

Before packaging, platform-specific whisper binaries must be staged under `packaging/whisper/...`.

For exact expected paths and host-local staging instructions, see `packaging/whisper/README.md`.

Release archives include the staged binary as:

```text
libexec/whisper/whisper-cli
```

Release archives also include licensing docs:

```text
LICENSE
THIRD_PARTY_NOTICES.md
```

## GitHub Actions release flow

`.github/workflows/release.yml` runs on version tags (`v*`) and manual dispatch.

The workflow builds and stages `whisper-cli` for:

- `linux/amd64`
- `linux/arm64`
- `darwin/amd64`
- `darwin/arm64`

It then runs GoReleaser to publish release archives and `checksums.txt`.

Release delivery is via **GitHub Releases assets** (not GitHub Packages):

- `voxclip_<version>_<os>_<arch>.tar.gz`
- `checksums.txt`

Each per-platform archive bundles both executables (`voxclip` and `libexec/whisper/whisper-cli`), so end-user installs do not compile `whisper.cpp` locally.

## Homebrew tap publishing

Releases automatically update the Homebrew formula in `fmueller/homebrew-tap` via GoReleaser's `brews:` stanza.

### Required secret

Add a GitHub secret named `HOMEBREW_TAP_TOKEN` to the `fmueller/voxclip` repository:

1. Create a fine-grained PAT at **GitHub Settings > Developer Settings > Fine-grained tokens**.
2. Scope it to `fmueller/homebrew-tap` with **Contents: Read and write** permission.
3. Add it as a repository secret in **Settings > Secrets and variables > Actions**.

Prerelease tags (containing a hyphen) skip the tap update (`skip_upload: auto`).
