# License Compatibility and Compliance Audit

Date: 2026-02-21

## Scope

This audit covers all dependencies declared in `go.mod` (direct and indirect), and identifies which dependencies are linked into the release CLI binary.

## Commands used

- `go list -m all`
- `go list -deps -f '{{with .Module}}{{.Path}} {{.Version}}{{end}}' ./cmd/voxclip | sort -u`
- `go mod download all`
- License files inspected from module cache under `$(go env GOPATH)/pkg/mod/...`.

## Compatibility result (against Apache-2.0 project license)

All dependencies in `go.mod` use permissive licenses that are compatible with Apache-2.0 distribution.

## Dependency license matrix

| Module | Version | License | Apache-2.0 compatibility | Notes |
|---|---:|---|---|---|
| github.com/schollz/progressbar/v3 | v3.19.0 | MIT | Compatible | Runtime dependency |
| github.com/spf13/cobra | v1.10.2 | Apache-2.0 | Compatible | Runtime dependency |
| github.com/stretchr/testify | v1.11.1 | MIT | Compatible | Test dependency |
| go.uber.org/zap | v1.27.1 | MIT | Compatible | Runtime dependency |
| golang.org/x/term | v0.40.0 | BSD-3-Clause | Compatible | Runtime dependency |
| github.com/davecgh/go-spew | v1.1.1 | ISC | Compatible | Test dependency |
| github.com/inconshreveable/mousetrap | v1.1.0 | Apache-2.0 | Compatible | Platform-specific dependency (not in Linux/macOS runtime list) |
| github.com/mitchellh/colorstring | v0.0.0-20190213212951-d06e56a500db | MIT | Compatible | Runtime dependency |
| github.com/pmezard/go-difflib | v1.0.0 | BSD-3-Clause | Compatible | Test dependency |
| github.com/rivo/uniseg | v0.4.7 | MIT | Compatible | Runtime dependency |
| github.com/spf13/pflag | v1.0.9 | BSD-3-Clause | Compatible | Runtime dependency |
| go.uber.org/multierr | v1.10.0 | MIT | Compatible | Runtime dependency |
| golang.org/x/sys | v0.41.0 | BSD-3-Clause | Compatible | Runtime dependency |
| gopkg.in/yaml.v3 | v3.0.1 | MIT + Apache-2.0 | Compatible | Test dependency; includes NOTICE file |

## Compliance assessment for Voxclip

### Current state

- Project has a primary Apache-2.0 `LICENSE` file.
- Prior to this update, release archives included `LICENSE` and `README.md` but no third-party notice/license summary.

### Actions taken in this change

- Added `THIRD_PARTY_NOTICES.md` with runtime and test/development dependency attributions and license identifiers.
- Updated `.goreleaser.yml` to include `THIRD_PARTY_NOTICES.md` in release archives.

### Practical release guidance

For binary releases, keep including:

1. `LICENSE` (project Apache-2.0 license)
2. `THIRD_PARTY_NOTICES.md` (third-party attribution summary)

This satisfies common notice-preservation expectations for MIT/BSD/ISC/Apache-style dependencies used by Voxclip.

## Caveat

This is a technical/compliance-oriented audit, not legal advice.
