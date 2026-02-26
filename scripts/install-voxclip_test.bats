#!/usr/bin/env bats

setup() {
  export _VOXCLIP_TESTING=1
  source "${BATS_TEST_DIRNAME}/install-voxclip.sh"
}

# detect_os tests

@test "detect_os returns linux or darwin" {
  result="$(detect_os)"
  [[ "$result" == "linux" || "$result" == "darwin" ]]
}

@test "detect_os returns linux on Linux" {
  if [[ "$(uname -s)" != "Linux" ]]; then
    skip "not a Linux runner"
  fi
  result="$(detect_os)"
  [[ "$result" == "linux" ]]
}

@test "detect_os returns darwin on macOS" {
  if [[ "$(uname -s)" != "Darwin" ]]; then
    skip "not a macOS runner"
  fi
  result="$(detect_os)"
  [[ "$result" == "darwin" ]]
}

# detect_arch tests

@test "detect_arch returns amd64 or arm64" {
  result="$(detect_arch)"
  [[ "$result" == "amd64" || "$result" == "arm64" ]]
}

@test "detect_arch returns amd64 on x86_64" {
  if [[ "$(uname -m)" != "x86_64" ]]; then
    skip "not an x86_64 runner"
  fi
  result="$(detect_arch)"
  [[ "$result" == "amd64" ]]
}

@test "detect_arch returns arm64 on aarch64" {
  local machine
  machine="$(uname -m)"
  if [[ "$machine" != "aarch64" && "$machine" != "arm64" ]]; then
    skip "not an arm64 runner"
  fi
  result="$(detect_arch)"
  [[ "$result" == "arm64" ]]
}

# sha256_file tests

@test "sha256_file produces a 64-character hex digest" {
  local tmp
  tmp="$(mktemp)"
  echo "test content" > "$tmp"
  result="$(sha256_file "$tmp")"
  rm -f "$tmp"
  [[ "${#result}" -eq 64 ]]
  [[ "$result" =~ ^[0-9a-f]+$ ]]
}

@test "sha256_file produces correct hash for known input" {
  local tmp
  tmp="$(mktemp)"
  printf 'hello\n' > "$tmp"
  result="$(sha256_file "$tmp")"
  rm -f "$tmp"
  [[ "$result" == "5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03" ]]
}

@test "sha256_file produces different hashes for different content" {
  local tmp1 tmp2
  tmp1="$(mktemp)"
  tmp2="$(mktemp)"
  echo "content a" > "$tmp1"
  echo "content b" > "$tmp2"
  hash1="$(sha256_file "$tmp1")"
  hash2="$(sha256_file "$tmp2")"
  rm -f "$tmp1" "$tmp2"
  [[ "$hash1" != "$hash2" ]]
}

# need_cmd tests

@test "need_cmd returns 0 for bash" {
  need_cmd bash
}

@test "need_cmd returns non-zero for nonexistent command" {
  ! need_cmd __no_such_command_voxclip_test__
}
