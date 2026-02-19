# Bundled whisper binaries

Place prebuilt whisper CLI binaries here before running GoReleaser.

For local host staging, run:

```bash
./scripts/build-whisper-cli.sh
```

For tagged releases, `.github/workflows/release.yml` builds and stages these binaries automatically.

Expected layout:

- `packaging/whisper/linux_amd64/whisper-cli`
- `packaging/whisper/linux_arm64/whisper-cli`
- `packaging/whisper/darwin_amd64/whisper-cli`
- `packaging/whisper/darwin_arm64/whisper-cli`

GoReleaser archives each file into `libexec/whisper/whisper-cli`.
