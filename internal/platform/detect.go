package platform

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type Runtime struct {
	OS   string
	Arch string
}

func CurrentRuntime() Runtime {
	return Runtime{
		OS:   runtime.GOOS,
		Arch: NormalizeArch(runtime.GOARCH),
	}
}

func NormalizeArch(arch string) string {
	switch arch {
	case "x86_64":
		return "amd64"
	case "aarch64":
		return "arm64"
	default:
		return arch
	}
}

func DefaultModelDirFor(goos, homeDir, xdgDataHome string) (string, error) {
	dataDir, err := defaultDataDirFor(goos, homeDir, xdgDataHome)
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, "models"), nil
}

func DefaultRecordingDirFor(goos, homeDir, xdgDataHome string) (string, error) {
	dataDir, err := defaultDataDirFor(goos, homeDir, xdgDataHome)
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, "recordings"), nil
}

func ResolveModelDir(override string) (string, error) {
	if override != "" {
		return filepath.Clean(override), nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home: %w", err)
	}

	return DefaultModelDirFor(runtime.GOOS, homeDir, os.Getenv("XDG_DATA_HOME"))
}

func ResolveRecordingDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home: %w", err)
	}

	return DefaultRecordingDirFor(runtime.GOOS, homeDir, os.Getenv("XDG_DATA_HOME"))
}

func defaultDataDirFor(goos, homeDir, xdgDataHome string) (string, error) {
	if homeDir == "" {
		return "", errors.New("home directory is empty")
	}

	switch goos {
	case "linux":
		if xdgDataHome != "" {
			return filepath.Join(xdgDataHome, "voxclip"), nil
		}
		return filepath.Join(homeDir, ".local", "share", "voxclip"), nil
	case "darwin":
		return filepath.Join(homeDir, "Library", "Application Support", "voxclip"), nil
	default:
		return "", fmt.Errorf("unsupported OS: %s", goos)
	}
}
