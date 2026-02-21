# Third-Party Notices

This project is licensed under Apache License 2.0. Voxclip depends on third-party libraries with permissive licenses that are compatible with Apache-2.0.

When redistributing Voxclip binaries, include this file alongside the project's `LICENSE` file.

## Runtime dependencies (linked into `voxclip` binary)

- `github.com/mitchellh/colorstring` (MIT)
- `github.com/rivo/uniseg` (MIT)
- `github.com/schollz/progressbar/v3` (MIT)
- `github.com/spf13/cobra` (Apache-2.0)
- `github.com/spf13/pflag` (BSD-3-Clause)
- `go.uber.org/multierr` (MIT)
- `go.uber.org/zap` (MIT)
- `golang.org/x/sys` (BSD-3-Clause)
- `golang.org/x/term` (BSD-3-Clause)

## Development and test-only dependencies

- `github.com/davecgh/go-spew` (ISC)
- `github.com/inconshreveable/mousetrap` (Apache-2.0)
- `github.com/pmezard/go-difflib` (BSD-3-Clause)
- `github.com/stretchr/testify` (MIT)
- `gopkg.in/yaml.v3` (MIT and Apache-2.0)

## Notes

- MIT, BSD-3-Clause, ISC, and Apache-2.0 are generally compatible with distribution in an Apache-2.0 project.
- Some licenses require preservation of copyright and license notices in source and/or binary redistributions. This file is provided to satisfy notice-preservation expectations.
- For any future dependency changes, rerun a license audit before release.
