# Third-Party Notices

This project is licensed under Apache License 2.0. Voxclip depends on third-party libraries and tools whose licenses are compatible with Apache-2.0. This file is included in release archives to satisfy notice-preservation requirements of those licenses.

## Bundled binary: whisper.cpp

Release archives include a prebuilt `whisper-cli` binary compiled from [whisper.cpp](https://github.com/ggml-org/whisper.cpp) (MIT).

## Runtime dependencies (compiled into `voxclip` binary)

- `github.com/inconshreveable/mousetrap` (Apache-2.0)
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
- `github.com/pmezard/go-difflib` (BSD-3-Clause)
- `github.com/stretchr/testify` (MIT)
- `gopkg.in/yaml.v3` (MIT and Apache-2.0)

## Test data assets

- FSDD (Free Spoken Digit Dataset) sample WAV files in `testdata/audio/fsdd/` (CC BY-SA 4.0)
  - Source: `https://github.com/Jakobovski/free-spoken-digit-dataset`
  - License: `https://creativecommons.org/licenses/by-sa/4.0/`

## Notes

- MIT, BSD-3-Clause, ISC, and Apache-2.0 are generally compatible with distribution in an Apache-2.0 project.
- Some licenses require preservation of copyright and license notices in source and binary redistributions.
- For dependency changes, CI checks licenses automatically via `go-licenses`.
