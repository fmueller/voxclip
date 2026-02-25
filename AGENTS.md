# AGENTS.md
Guidance for coding agents working in the Voxclip repository.

## Scope and intent
- This is a Go CLI repository: `github.com/fmueller/voxclip`.
- Main executable: `./cmd/voxclip`.
- Internal packages: `./internal/...`.
- Keep changes small and behavior-preserving unless explicitly requested.

## Source-of-truth files
- Commands and local workflows: `Taskfile.yml`.
- CI checks and required validations: `.github/workflows/ci.yml`.
- Release pipeline: `.github/workflows/release.yml` and `.goreleaser.yml`.
- Release process and packaging docs: `RELEASING.md` and `packaging/whisper/README.md`.
- Product behavior and user docs: `README.md`.

## Toolchain and environment
- Go version: `1.26` (`go.mod`).
- `task` is optional convenience; Go commands are the source of truth.
- Prefer commands that mirror CI.

## Build, lint, and test commands
Use these defaults unless a task requires otherwise.

### Build
- Build CLI binary: `go build ./cmd/voxclip`
- Task wrapper: `task build`

### Formatting and static checks
- Check formatting (CI): `gofmt -l .`
- Apply formatting: `gofmt -w .`
- Vet: `go vet ./...`

### Unit tests
- Run all unit tests: `go test ./...`
- Task wrapper: `task test`

### Integration tests
- Run integration-tag tests: `go test -tags=integration ./...`
- Task wrapper: `task test:integration`

### Run a single test
- Single test in one package:
  `go test ./internal/cli -run '^TestTranscribeCommandSkipsCopyForBlankTranscript$'`
- Single test by pattern:
  `go test ./internal/cli -run '^TestTranscribeCommand'`
- Single integration test:
  `go test -tags=integration ./internal/download -run '^TestDownloadFileEndToEndWithFixtureServer$'`
- Single test with verbose output:
  `go test -v ./cmd/voxclip -run '^TestHelpHintTarget$'`

### Helpful test variants
- Disable cache while iterating:
  `go test ./internal/cli -run '^TestName$' -count=1`
- Run one package only:
  `go test ./internal/audio`

### CLI smoke checks used in CI
- `go run ./cmd/voxclip --help`
- `go run ./cmd/voxclip setup --help`
- `go run ./cmd/voxclip record --help`
- `go run ./cmd/voxclip transcribe --help`
- `go run ./cmd/voxclip devices --help`

### Mutation tests
- Run mutation testing: `gremlins unleash`
- Task wrapper: `task test:mutate`
- With quality gate (used in CI): `gremlins unleash --threshold-efficacy 0.25`
- Install gremlins: `go install github.com/go-gremlins/gremlins/cmd/gremlins@v0.6.0`

### Installer script checks used in CI
- Syntax check: `bash -n scripts/install-voxclip.sh`
- Help command: `./scripts/install-voxclip.sh --help`

### Release-related commands
- Build bundled whisper binary for current host: `./scripts/build-whisper-cli.sh`
- Task wrapper: `task whisper:build`
- Validate GoReleaser config: `goreleaser check` (or `task release:check`)
- Dry snapshot release: `goreleaser release --snapshot --clean`

## Coding style guidelines
Follow existing conventions and keep CLI UX stable.

### Formatting and structure
- Always run `gofmt` on changed Go files.
- Keep functions focused and prefer early returns for error paths.
- Avoid unnecessary abstractions; match current package boundaries.

### Imports
- Let `gofmt` handle import order.
- Keep stdlib imports separated from non-stdlib imports.
- Prefer module-local imports (`github.com/fmueller/voxclip/internal/...`) for internal reuse.

### Types and API shape
- Exported names: `PascalCase`; unexported names: `camelCase`.
- Sentinel errors should use `ErrXxx` naming (for example `ErrUnavailable`).
- Exported constructors should use `NewXxx`.
- Use explicit config structs (for example `record.Config`, `download.Options`).
- Prefer typed structs for core flows instead of `map[string]any`.

### Naming conventions
- Package names are short, lowercase, and noun-like.
- CLI command builders use `newXCmd` for unexported command constructors.
- Booleans should read clearly (`noProgress`, `autoDownload`, `copyEmpty`).
- Error text should be lowercase and not end with punctuation.

### Error handling
- Return errors instead of panicking.
- Wrap underlying errors with context using `%w`.
  Example: `fmt.Errorf("download model %q: %w", name, err)`
- Use `errors.Is` and `errors.As` for known error branches.
- Keep errors actionable for CLI users.

### Context and I/O
- Pass `context.Context` as the first parameter when operations are cancelable.
- Use `exec.CommandContext` and context-bound HTTP requests.
- Prefer `filepath` (not `path`) for filesystem operations.
- Normalize user-provided paths with `filepath.Clean` when appropriate.

### Logging and user output
- Use structured logging with `zap` for diagnostics.
- Write user-facing command output via command writers (`cmd.OutOrStdout()` / stderr).
- Preserve transcript/output behavior in existing command flows.
- Keep no-speech behavior aligned with `blankAudioToken` handling.

### Testing conventions
- Use standard `testing` + `github.com/stretchr/testify/require`.
- Use `t.Parallel()` for unit tests when safe.
- Mark integration tests with `//go:build integration`.
- Use `t.TempDir()` for temporary filesystem fixtures.
- Keep tests deterministic; avoid live network unless integration-scoped.

### Dependency and platform practices
- Prefer the standard library first; add dependencies only when justified.
- Keep Linux/macOS differences explicit and runtime-guarded.
- Preserve backend selection behavior (`auto`, `pw-record`, `arecord`, `ffmpeg`).

## Change checklist for agents
- Run `gofmt -w` on edited Go files.
- Run `go vet ./...` for non-trivial changes.
- Run targeted tests for touched packages.
- Run full `go test ./...` before handing off broad changes.
- If integration paths changed, run `go test -tags=integration ./...`.
- Update `README.md` when CLI flags/commands/behavior change.
- Update `RELEASING.md` (and `packaging/whisper/README.md` when needed) when release flow, packaging layout, or required secrets change.
- Update `CHANGELOG.md` for user-visible behavior changes; keep entries user-facing and skip internal-only maintenance noise.

## Agent do/don'ts (PR + commits)
- Do keep branches and PRs focused on one logical change.
- Do include a clear PR summary with why, what changed, and test evidence.
- Do keep commits small and descriptive, using imperative commit subjects.
- Do use conventional commit messages.
- Do mention user-visible CLI changes and migration notes in the PR body.
- Don't mix unrelated refactors or formatting-only churn into feature/fix PRs.
- Don't rewrite the shared branch history unless explicitly requested.
- Don't bypass CI-equivalent checks before asking for review.

## Notes on repository behavior
- Default CLI flow is `record -> transcribe -> copy`.
- Blank transcripts are a deliberate no-speech case.
- Clipboard failure is often non-fatal in default flow; preserve that UX.
- There is intentionally no `voxclip install` command.
