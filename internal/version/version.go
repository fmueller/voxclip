package version

import (
	"os/exec"
	"strings"
)

var (
	Version = "1.0.0"
	Commit  = "unknown"
	Date    = "unknown"
)

// Resolve returns the full version string, appending a git-derived suffix
// when the binary is run from inside a git repository whose HEAD is not on
// a release tag.
func Resolve() string {
	return resolveVersion(Version, runGit)
}

func resolveVersion(base string, git func(...string) (string, error)) string {
	if base == "" {
		base = "0.0.0"
	}

	suffix := computeGitSuffix(base, git)
	if suffix == "" {
		return base
	}
	return base + "-" + suffix
}

func computeGitSuffix(base string, git func(...string) (string, error)) string {
	if _, err := git("rev-parse", "--git-dir"); err != nil {
		return ""
	}

	if _, err := git("describe", "--tags", "--exact-match"); err == nil {
		return ""
	}

	desc, err := git("describe", "--tags", "--dirty", "--always")
	if err != nil {
		return ""
	}

	prefix := "v" + base + "-"
	if strings.HasPrefix(desc, prefix) {
		return strings.TrimPrefix(desc, prefix)
	}

	return desc
}

func runGit(args ...string) (string, error) {
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
