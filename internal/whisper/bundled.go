package whisper

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"go.uber.org/zap"
)

type BundledEngine struct {
	Executable string
	Logger     *zap.Logger
}

func NewBundledEngine(logger *zap.Logger) (*BundledEngine, error) {
	if logger == nil {
		logger = zap.NewNop()
	}

	if override := strings.TrimSpace(os.Getenv("VOXCLIP_WHISPER_PATH")); override != "" {
		if err := ensureExecutable(override); err != nil {
			return nil, fmt.Errorf("VOXCLIP_WHISPER_PATH is not executable: %w", err)
		}
		return &BundledEngine{Executable: override, Logger: logger}, nil
	}

	voxclipExe, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("resolve voxclip executable path: %w", err)
	}

	whisperExe, err := ResolveBundledEnginePath(voxclipExe)
	if err != nil {
		return nil, err
	}

	return &BundledEngine{Executable: whisperExe, Logger: logger}, nil
}

func ResolveBundledEnginePath(voxclipExecutable string) (string, error) {
	for _, candidate := range EnginePathCandidates(voxclipExecutable) {
		if err := ensureExecutable(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("bundled whisper engine not found near %s; reinstall Voxclip from official release, expected at ../libexec/whisper/%s", voxclipExecutable, engineBinaryName())
}

func EnginePathCandidates(voxclipExecutable string) []string {
	binDir := filepath.Dir(voxclipExecutable)
	engineName := engineBinaryName()
	hostTarget := fmt.Sprintf("%s_%s", runtime.GOOS, normalizeArch(runtime.GOARCH))

	return []string{
		filepath.Join(binDir, "..", "libexec", "whisper", engineName),
		filepath.Join(binDir, "libexec", "whisper", engineName),
		filepath.Join(binDir, "packaging", "whisper", hostTarget, engineName),
		filepath.Join(binDir, engineName),
	}
}

func (b *BundledEngine) Transcribe(ctx context.Context, req TranscriptionRequest) (string, error) {
	if strings.TrimSpace(req.AudioPath) == "" {
		return "", errors.New("audio path is required")
	}
	if strings.TrimSpace(req.ModelPath) == "" {
		return "", errors.New("model path is required")
	}

	if err := ensureExecutable(b.Executable); err != nil {
		return "", fmt.Errorf("bundled whisper engine missing or not executable: %w", err)
	}

	outBase := filepath.Join(os.TempDir(), fmt.Sprintf("voxclip-%d", time.Now().UnixNano()))
	txtOut := outBase + ".txt"

	args := []string{"-m", req.ModelPath, "-f", req.AudioPath, "-nt", "-otxt", "-of", outBase}
	lang := strings.TrimSpace(req.Language)
	if lang != "" && lang != "auto" {
		args = append(args, "-l", lang)
	}

	cmd := exec.CommandContext(ctx, b.Executable, args...)
	var stderr bytes.Buffer
	cmd.Stdout = ioDiscard{}
	cmd.Stderr = &stderr

	b.Logger.Debug("running whisper engine", zap.String("engine", b.Executable), zap.Strings("args", args))
	if err := cmd.Run(); err != nil {
		errText := strings.TrimSpace(stderr.String())
		if isMissingSharedLibraryError(errText) {
			return "", fmt.Errorf("bundled whisper engine at %s is missing required shared libraries (%s); reinstall Voxclip from an official release or rebuild whisper-cli with BUILD_SHARED_LIBS=OFF", b.Executable, errText)
		}
		if isIllegalInstructionError(errText) || isIllegalInstructionError(err.Error()) {
			return "", fmt.Errorf("bundled whisper engine crashed with an illegal CPU instruction; " +
				"your CPU may lack required instruction set extensions; " +
				"set VOXCLIP_WHISPER_PATH to a whisper-cli binary built for your CPU")
		}
		return "", fmt.Errorf("whisper transcribe failed: %w (%s)", err, errText)
	}

	defer os.Remove(txtOut)
	content, err := os.ReadFile(txtOut)
	if err != nil {
		return "", fmt.Errorf("read whisper output: %w", err)
	}

	return strings.TrimSpace(string(content)), nil
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) {
	return len(p), nil
}

func engineBinaryName() string {
	if runtime.GOOS == "windows" {
		return "whisper-cli.exe"
	}
	return "whisper-cli"
}

func ensureExecutable(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("%s is a directory", path)
	}
	if runtime.GOOS != "windows" && info.Mode()&0o111 == 0 {
		return fmt.Errorf("%s is not executable", path)
	}
	return nil
}

func isMissingSharedLibraryError(stderr string) bool {
	value := strings.ToLower(strings.TrimSpace(stderr))
	if value == "" {
		return false
	}

	patterns := []string{
		"error while loading shared libraries",
		"cannot open shared object file",
		"dyld: library not loaded",
		"image not found",
	}

	for _, pattern := range patterns {
		if strings.Contains(value, pattern) {
			return true
		}
	}

	return false
}

func isIllegalInstructionError(stderr string) bool {
	return strings.Contains(strings.ToLower(stderr), "illegal instruction")
}

func normalizeArch(arch string) string {
	switch arch {
	case "x86_64":
		return "amd64"
	case "aarch64":
		return "arm64"
	default:
		return arch
	}
}
