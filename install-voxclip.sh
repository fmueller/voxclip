#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<HELP
Usage:
  install-voxclip.sh [--full-install]

Options:
  --full-install   Install launcher and run 'voxclip install'
  -h, --help       Show this help
HELP
}

FULL_INSTALL=0
while [[ $# -gt 0 ]]; do
  case "$1" in
    --full-install)
      FULL_INSTALL=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

mkdir -p "$HOME/.local/bin"

cat > "$HOME/.local/bin/voxclip" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

# voxclip: cross-platform (Linux/macOS) self-bootstrapping voice-to-text via whisper.cpp
#
# Commands:
#   voxclip install   # deps + clone + build + model
#   voxclip           # default: run (auto-installs if needed)
#   voxclip run       # record -> transcribe -> clipboard
#   voxclip update    # git pull + rebuild
#   voxclip devices   # show input devices / hints
#   voxclip uninstall # remove installed files under ROOT (keeps deps)
#
# Config file (optional):
#   ~/.config/voxclip/config.env
# Example:
#   WHISPER_MODEL=small
#   WHISPER_LANGUAGE=auto
#   WHISPER_DURATION_SEC=0
#   WHISPER_THREADS=8
#   WHISPER_AVF_INPUT=:0

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
is_macos() { [[ "$OS" == "darwin" ]]; }
is_linux() { [[ "$OS" == "linux" ]]; }

need_cmd() { command -v "$1" >/dev/null 2>&1; }

REC_WAV_PATH=""
REC_PID=""

has_tty_input() {
  [[ -r /dev/tty ]] || [[ -t 0 ]]
}

wait_for_enter() {
  local prompt="$1"
  printf '%s\n' "$prompt" >&2

  if [[ -r /dev/tty ]]; then
    read -r < /dev/tty || return 1
    return 0
  fi

  if [[ -t 0 ]]; then
    read -r || return 1
    return 0
  fi

  return 1
}

is_valid_pid() {
  local pid="${1:-}"
  [[ "$pid" =~ ^[0-9]+$ ]] && [[ "$pid" -gt 0 ]]
}

CONFIG_FILE="${VOXCLIP_CONFIG:-$HOME/.config/voxclip/config.env}"
if [[ -f "$CONFIG_FILE" ]]; then
  # shellcheck disable=SC1090
  source "$CONFIG_FILE"
fi

ROOT="${VOXCLIP_ROOT:-$HOME/.local/share/voxclip}"
REPO_DIR="$ROOT/whisper.cpp"
REC_DIR="${VOXCLIP_REC_DIR:-$ROOT/recordings}"

MODEL_NAME="${WHISPER_MODEL:-small}"          # multilingual: tiny|base|small|medium|large-v3|...
LANGUAGE="${WHISPER_LANGUAGE:-auto}"          # auto|en|de
DURATION_SEC="${WHISPER_DURATION_SEC:-0}"     # 0 = press Enter to stop
THREADS="${WHISPER_THREADS:-0}"               # 0 = auto

# macOS recording input (ffmpeg avfoundation)
AVF_INPUT="${WHISPER_AVF_INPUT:-:0}"

# Linux: try pulse default then alsa default (can override via flags)
LINUX_FFMPEG_FMT="${WHISPER_LINUX_FFMPEG_FMT:-}"
LINUX_FFMPEG_IN="${WHISPER_LINUX_FFMPEG_IN:-}"

auto_threads() {
  if [[ "$THREADS" != "0" ]]; then
    echo "$THREADS"
    return
  fi
  if is_macos; then
    sysctl -n hw.ncpu 2>/dev/null || echo 4
  else
    nproc 2>/dev/null || echo 4
  fi
}

usage() {
  cat <<HELP
Usage:
  voxclip [command] [options]

Commands:
  install                 Install deps, clone/build whisper.cpp, download model
  run                     Record -> transcribe -> clipboard (default)
  update                  git pull + rebuild
  devices                 Show recording device info (macOS lists AVFoundation devices)
  uninstall               Remove installed files under ROOT
  help                    Show this help

Run options:
  -l, --lang  auto|en|de
  -m, --model tiny|base|small|medium|large-v3|...
  -d, --duration SEC      Record SEC seconds (0 = press Enter to stop)
  -t, --threads N
  --root PATH             Override installation root

macOS options:
  --avf ":0"              AVFoundation input (audio-only). Use: voxclip devices

Linux options:
  --ffmpeg-fmt pulse|alsa         Override input format
  --ffmpeg-in  default|<device>   Override input device

Config file (optional):
  $CONFIG_FILE
HELP
}

ensure_config_dir() {
  mkdir -p "$(dirname "$CONFIG_FILE")"
}

write_default_config_if_missing() {
  ensure_config_dir
  if [[ ! -f "$CONFIG_FILE" ]]; then
    cat > "$CONFIG_FILE" <<CFG
# voxclip config
# Edit these defaults as you like:
WHISPER_MODEL=small
WHISPER_LANGUAGE=auto
WHISPER_DURATION_SEC=0
# WHISPER_THREADS=8

# macOS mic selection for ffmpeg avfoundation (audio-only):
WHISPER_AVF_INPUT=:0

# Linux overrides (optional):
# WHISPER_LINUX_FFMPEG_FMT=pulse
# WHISPER_LINUX_FFMPEG_IN=default
CFG
  fi
}

install_deps_linux() {
  declare -A pkgs=()
  need_cmd git || pkgs[git]=1
  need_cmd cmake || pkgs[cmake]=1
  need_cmd make || pkgs[build-essential]=1
  need_cmd g++ || pkgs[build-essential]=1
  need_cmd ffmpeg || pkgs[ffmpeg]=1

  # Clipboard: prefer wl-copy on Wayland; xclip on X11; install both if none exist.
  if ! need_cmd wl-copy && ! need_cmd xclip; then
    pkgs[wl-clipboard]=1
    pkgs[xclip]=1
  fi

  if (( ${#pkgs[@]} > 0 )); then
    if ! need_cmd sudo; then
      echo "Missing dependencies and 'sudo' not found. Install manually:" >&2
      printf '  %s\n' "${!pkgs[@]}" >&2
      exit 1
    fi
    echo "Installing packages: ${!pkgs[@]}"
    sudo apt-get update -y
    sudo apt-get install -y "${!pkgs[@]}"
  fi
}

install_deps_macos() {
  # Xcode CLT
  if ! xcode-select -p >/dev/null 2>&1; then
    echo "Xcode Command Line Tools not found. Triggering install..."
    xcode-select --install || true
    echo "Finish the GUI install if prompted, then re-run: voxclip install"
    # We still proceed; build will fail until installed.
  fi

  if ! need_cmd brew; then
    echo "Homebrew not found. Installing (may prompt for password)..."
    NONINTERACTIVE=1 /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    if [[ -x /opt/homebrew/bin/brew ]]; then
      eval "$(/opt/homebrew/bin/brew shellenv)"
    elif [[ -x /usr/local/bin/brew ]]; then
      eval "$(/usr/local/bin/brew shellenv)"
    fi
  fi

  for p in git cmake ffmpeg; do
    if ! need_cmd "$p"; then
      echo "Installing $p via brew..."
      brew install "$p"
    fi
  done
}

ensure_deps() {
  if is_linux; then
    install_deps_linux
  elif is_macos; then
    install_deps_macos
  else
    echo "Unsupported OS: $OS" >&2
    exit 1
  fi
}

ensure_repo() {
  mkdir -p "$ROOT"
  if [[ ! -d "$REPO_DIR/.git" ]]; then
    echo "Cloning whisper.cpp into: $REPO_DIR"
    git clone --depth 1 https://github.com/ggml-org/whisper.cpp.git "$REPO_DIR"
  fi
}

build_whisper() {
  local threads
  threads="$(auto_threads)"
  echo "Building whisper.cpp (Release) with $threads threads..."
  cmake -S "$REPO_DIR" -B "$REPO_DIR/build" -DCMAKE_BUILD_TYPE=Release
  cmake --build "$REPO_DIR/build" -j "$threads" --config Release --target whisper-cli 2>/dev/null || \
    cmake --build "$REPO_DIR/build" -j "$threads" --config Release
}

ensure_model() {
  local model_file="$REPO_DIR/models/ggml-${MODEL_NAME}.bin"
  if [[ ! -f "$model_file" ]]; then
    echo "Downloading model: $MODEL_NAME"
    (cd "$REPO_DIR" && sh ./models/download-ggml-model.sh "$MODEL_NAME")
  fi
}

is_installed() {
  [[ -x "$REPO_DIR/build/bin/whisper-cli" ]]
}

pick_clip_cmd() {
  if is_macos; then
    echo "pbcopy"
  else
    if need_cmd wl-copy; then
      echo "wl-copy"
    elif need_cmd xclip; then
      echo "xclip -selection clipboard"
    else
      echo ""
    fi
  fi
}

cmd_install() {
  write_default_config_if_missing
  ensure_deps
  ensure_repo
  build_whisper
  ensure_model
  echo "Install complete."
  echo "Run: voxclip   (or voxclip run)"
}

cmd_update() {
  ensure_deps
  ensure_repo
  echo "Updating repo..."
  (cd "$REPO_DIR" && git pull --rebase)
  build_whisper
  ensure_model
  echo "Update complete."
}

cmd_uninstall() {
  echo "Removing: $ROOT"
  rm -rf "$ROOT"
  echo "Uninstalled voxclip assets. (Dependency packages not removed.)"
}

cmd_devices() {
  if is_macos; then
    echo "AVFoundation devices (find your mic input index):"
    ffmpeg -hide_banner -f avfoundation -list_devices true -i "" 2>&1 | sed -n '1,200p'
    echo
    echo "Then use, for example: voxclip run --avf \":1\""
  else
    echo "Linux recording hints:"
    echo "  pactl list short sources   # PulseAudio/PipeWire sources"
    echo "  arecord -L                 # ALSA devices"
    echo
    echo "You can override input like:"
    echo "  voxclip run --ffmpeg-fmt pulse --ffmpeg-in default"
  fi
}

record_audio_macos() {
  mkdir -p "$REC_DIR"
  local ts wav pid
  ts="$(date +%Y%m%d-%H%M%S)"
  wav="$REC_DIR/recording-$ts.wav"

  wait_for_enter "Press Enter to start recording..." || {
    echo "Interactive input unavailable. Use --duration SEC for non-interactive runs." >&2
    exit 1
  }

  echo "Recording via AVFoundation (input=$AVF_INPUT)..." >&2
  if [[ "$DURATION_SEC" != "0" ]]; then
    ffmpeg -nostdin -hide_banner -loglevel error -y \
      -f avfoundation -i "$AVF_INPUT" \
      -t "$DURATION_SEC" \
      -ac 1 -ar 16000 -c:a pcm_s16le \
      "$wav" &
  else
    ffmpeg -nostdin -hide_banner -loglevel error -y \
      -f avfoundation -i "$AVF_INPUT" \
      -ac 1 -ar 16000 -c:a pcm_s16le \
      "$wav" &
  fi

  pid=$!
  REC_PID="$pid"
  sleep 0.3
  if ! is_valid_pid "$pid" || ! kill -0 "$pid" >/dev/null 2>&1; then
    wait "$pid" >/dev/null 2>&1 || true
    echo "Recording failed. Run: voxclip devices, then set --avf \":<index>\"" >&2
    exit 1
  fi

  if [[ "$DURATION_SEC" == "0" ]]; then
    echo "Recording... press Enter to stop." >&2
  else
    echo "Recording for $DURATION_SEC seconds..." >&2
  fi

  if [[ "$DURATION_SEC" == "0" ]]; then
    wait_for_enter "Press Enter to stop recording..." || {
      echo "Interactive input unavailable. Use --duration SEC for non-interactive runs." >&2
      exit 1
    }
    kill -INT "$pid" >/dev/null 2>&1 || true
  fi
  wait "$pid" >/dev/null 2>&1 || true
  REC_WAV_PATH="$wav"
}

start_recording_linux() {
  local wav="$1" fmt="$2" input="$3"
  REC_PID=""

  if [[ "$DURATION_SEC" != "0" ]]; then
    ffmpeg -nostdin -hide_banner -loglevel error -y \
      -f "$fmt" -i "$input" \
      -t "$DURATION_SEC" \
      -ac 1 -ar 16000 -c:a pcm_s16le \
      "$wav" &
  else
    ffmpeg -nostdin -hide_banner -loglevel error -y \
      -f "$fmt" -i "$input" \
      -ac 1 -ar 16000 -c:a pcm_s16le \
      "$wav" &
  fi
  REC_PID="$!"
}

record_audio_linux() {
  mkdir -p "$REC_DIR"
  local ts wav pid fmt input
  ts="$(date +%Y%m%d-%H%M%S)"
  wav="$REC_DIR/recording-$ts.wav"

  wait_for_enter "Press Enter to start recording..." || {
    echo "Interactive input unavailable. Use --duration SEC for non-interactive runs." >&2
    exit 1
  }

  if [[ -n "$LINUX_FFMPEG_FMT" && -n "$LINUX_FFMPEG_IN" ]]; then
    fmt="$LINUX_FFMPEG_FMT"
    input="$LINUX_FFMPEG_IN"
    echo "Recording via ffmpeg (fmt=$fmt, input=$input)..." >&2
    start_recording_linux "$wav" "$fmt" "$input"
    pid="$REC_PID"
  else
    # Auto: try PulseAudio/PipeWire first, then ALSA.
    echo "Recording via ffmpeg (trying pulse default, then alsa default)..." >&2
    start_recording_linux "$wav" pulse default
    pid="$REC_PID"
    sleep 0.2
    if ! is_valid_pid "$pid" || ! kill -0 "$pid" >/dev/null 2>&1; then
      start_recording_linux "$wav" alsa default
      pid="$REC_PID"
    fi
  fi

  sleep 0.3
  if ! is_valid_pid "$pid" || ! kill -0 "$pid" >/dev/null 2>&1; then
    if is_valid_pid "$pid"; then
      wait "$pid" >/dev/null 2>&1 || true
    fi
    echo "Recording failed. Try: voxclip devices (then override --ffmpeg-fmt/--ffmpeg-in)" >&2
    exit 1
  fi

  if [[ "$DURATION_SEC" == "0" ]]; then
    echo "Recording... press Enter to stop." >&2
  else
    echo "Recording for $DURATION_SEC seconds..." >&2
  fi

  if [[ "$DURATION_SEC" == "0" ]]; then
    wait_for_enter "Press Enter to stop recording..." || {
      echo "Interactive input unavailable. Use --duration SEC for non-interactive runs." >&2
      exit 1
    }
    kill -INT "$pid" >/dev/null 2>&1 || true
  fi
  wait "$pid" >/dev/null 2>&1 || true
  REC_WAV_PATH="$wav"
}

transcribe_and_copy() {
  local wav="$1"
  local threads
  threads="$(auto_threads)"

  local bin="$REPO_DIR/build/bin/whisper-cli"
  local model="$REPO_DIR/models/ggml-${MODEL_NAME}.bin"
  local out_base="${wav%.wav}"

  [[ -x "$bin" ]] || { echo "Missing binary: $bin" >&2; exit 1; }
  [[ -f "$model" ]] || { echo "Missing model: $model" >&2; exit 1; }

  echo "Transcribing (lang=$LANGUAGE, model=$MODEL_NAME, threads=$threads)..."
  "$bin" -m "$model" -f "$wav" -t "$threads" -l "$LANGUAGE" -nt -otxt -of "$out_base" >/dev/null

  local txt="${out_base}.txt"
  [[ -f "$txt" ]] || { echo "Transcript not found: $txt" >&2; exit 1; }

  echo
  echo "----- TRANSCRIPT -----"
  cat "$txt"
  echo "----------------------"
  echo

  local clip
  clip="$(pick_clip_cmd)"
  if [[ -z "$clip" ]]; then
    echo "No clipboard tool found. Transcript saved at: $txt"
    return
  fi

  if is_macos; then
    pbcopy < "$txt"
  else
    # shellcheck disable=SC2086
    eval $clip \< "$txt"
  fi
  echo "Copied to clipboard [OK]"
}

cmd_run() {
  if [[ "$DURATION_SEC" == "0" ]] && ! has_tty_input; then
    echo "Interactive mode requires a terminal input device." >&2
    echo "Run with --duration SEC for non-interactive execution (example: voxclip run -d 6)." >&2
    exit 1
  fi

  # Lazy install if needed
  if ! is_installed; then
    echo "Not installed yet -> running install first..."
    cmd_install
  else
    # Ensure model exists if user changed MODEL_NAME
    ensure_model
  fi

  mkdir -p "$REC_DIR"

  if is_macos; then
    record_audio_macos
  else
    record_audio_linux
  fi
  if [[ -z "$REC_WAV_PATH" ]]; then
    echo "Recording did not produce an output file path." >&2
    exit 1
  fi

  transcribe_and_copy "$REC_WAV_PATH"
}

# ------------------ argument parsing ------------------
CMD="${1:-run}"
shift || true

# parse common flags for run (allow flags after "run" or directly)
parse_run_flags() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
      -l|--lang) LANGUAGE="$2"; shift 2;;
      -m|--model) MODEL_NAME="$2"; shift 2;;
      -d|--duration) DURATION_SEC="$2"; shift 2;;
      -t|--threads) THREADS="$2"; shift 2;;
      --root)
        ROOT="$2"
        REPO_DIR="$ROOT/whisper.cpp"
        REC_DIR="$ROOT/recordings"
        shift 2
        ;;
      --avf) AVF_INPUT="$2"; shift 2;;
      --ffmpeg-fmt) LINUX_FFMPEG_FMT="$2"; shift 2;;
      --ffmpeg-in)  LINUX_FFMPEG_IN="$2"; shift 2;;
      -h|--help) usage; exit 0;;
      *) echo "Unknown option: $1" >&2; usage; exit 2;;
    esac
  done
}

case "$CMD" in
  help|-h|--help) usage;;
  install) cmd_install;;
  update) cmd_update;;
  uninstall) cmd_uninstall;;
  devices) cmd_devices;;
  run) parse_run_flags "$@"; cmd_run;;
  *)
    # If the user omitted "run", treat the first token as a flag set for run.
    # Example: voxclip --lang de
    if [[ "$CMD" == --* || "$CMD" == -* ]]; then
      # re-assemble args: include CMD back as first flag
      set -- "$CMD" "$@"
      parse_run_flags "$@"
      cmd_run
    else
      echo "Unknown command: $CMD" >&2
      usage
      exit 2
    fi
    ;;
esac
EOF

chmod +x "$HOME/.local/bin/voxclip"

echo "Installed: $HOME/.local/bin/voxclip"
echo "If needed, add to PATH:"
echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
echo

if [[ "$FULL_INSTALL" == "1" ]]; then
  echo "Running full install (deps/build/model)..."
  "$HOME/.local/bin/voxclip" install
  echo
  echo "Full install complete. Next step:"
  echo "  voxclip run"
else
  echo "No auto-run performed. Next steps:"
  echo "  voxclip run"
  echo "  voxclip run -d 6"
fi
