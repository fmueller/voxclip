package version

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func fakeGit(exactMatch string, describe string, exactErr, descErr error) func(...string) (string, error) {
	return func(args ...string) (string, error) {
		if len(args) == 0 {
			return "", fmt.Errorf("no args")
		}
		switch args[0] {
		case "rev-parse":
			return ".git", nil
		case "describe":
			for _, a := range args {
				if a == "--exact-match" {
					return exactMatch, exactErr
				}
			}
			return describe, descErr
		default:
			return "", fmt.Errorf("unexpected git subcommand %q", args[0])
		}
	}
}

func fakeGitNotARepo() func(...string) (string, error) {
	return func(args ...string) (string, error) {
		return "", fmt.Errorf("not a git repository")
	}
}

func TestResolveVersion_TaggedRelease(t *testing.T) {
	t.Parallel()
	git := fakeGit("v1.0.0", "", nil, nil)
	got := resolveVersion("1.0.0", git)
	require.Equal(t, "1.0.0", got)
}

func TestResolveVersion_CommitsAfterTag(t *testing.T) {
	t.Parallel()
	git := fakeGit("", "v1.0.0-3-gabcdef", fmt.Errorf("no tag"), nil)
	got := resolveVersion("1.0.0", git)
	require.Equal(t, "1.0.0-3-gabcdef", got)
}

func TestResolveVersion_DirtyWorkingTree(t *testing.T) {
	t.Parallel()
	git := fakeGit("", "v1.0.0-3-gabcdef-dirty", fmt.Errorf("no tag"), nil)
	got := resolveVersion("1.0.0", git)
	require.Equal(t, "1.0.0-3-gabcdef-dirty", got)
}

func TestResolveVersion_NoTags(t *testing.T) {
	t.Parallel()
	git := fakeGit("", "abcdef", fmt.Errorf("no tag"), nil)
	got := resolveVersion("1.0.0", git)
	require.Equal(t, "1.0.0-abcdef", got)
}

func TestResolveVersion_NotAGitRepo(t *testing.T) {
	t.Parallel()
	got := resolveVersion("1.0.0", fakeGitNotARepo())
	require.Equal(t, "1.0.0", got)
}

func TestResolveVersion_EmptyBaseFallsBackToZero(t *testing.T) {
	t.Parallel()
	got := resolveVersion("", fakeGitNotARepo())
	require.Equal(t, "0.0.0", got)
}

func TestResolveVersion_DescribeFails(t *testing.T) {
	t.Parallel()
	git := fakeGit("", "", fmt.Errorf("no tag"), fmt.Errorf("describe failed"))
	got := resolveVersion("1.0.0", git)
	require.Equal(t, "1.0.0", got)
}

func TestResolveVersion_DirtyNoTags(t *testing.T) {
	t.Parallel()
	git := fakeGit("", "abcdef-dirty", fmt.Errorf("no tag"), nil)
	got := resolveVersion("1.0.0", git)
	require.Equal(t, "1.0.0-abcdef-dirty", got)
}
