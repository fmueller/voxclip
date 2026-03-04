package version

import (
	"os/exec"
	"regexp"
	"strings"
)

var (
	Version = "1.0.0"
	Commit  = "unknown"
	Date    = "unknown"
)

// versionTagPrefix matches a semver-like tag prefix in git describe output,
// e.g. "v1.0.1-" in "v1.0.1-3-gabcdef".
var versionTagPrefix = regexp.MustCompile(`^v\d+\.\d+\.\d+-`)

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

	// Strip any semver tag prefix (e.g. "v1.2.3-") so that a describe
	// result from a newer tag doesn't produce "1.0.0-v1.0.1-3-gabcdef".
	if loc := versionTagPrefix.FindStringIndex(desc); loc != nil {
		return desc[loc[1]:]
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
