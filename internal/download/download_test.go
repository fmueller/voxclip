package download

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseChecksumByFilename(t *testing.T) {
	t.Parallel()

	content := []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa  foo.tar.gz\n" +
		"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb  checksums.txt\n")

	parsed, err := ParseChecksum(content, "foo.tar.gz")
	require.NoError(t, err)
	require.Equal(t, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", parsed)
}

func TestVerifyFileChecksum(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "payload.bin")
	payload := []byte("voxclip")
	require.NoError(t, os.WriteFile(path, payload, 0o644))

	sum := sha256.Sum256(payload)
	require.NoError(t, VerifyFileChecksum(path, hex.EncodeToString(sum[:])))
	require.Error(t, VerifyFileChecksum(path, "deadbeef"))
}

func TestDownloadFileWithChecksumURL(t *testing.T) {
	t.Parallel()

	payload := []byte("hello-world")
	sum := sha256.Sum256(payload)
	sumHex := hex.EncodeToString(sum[:])

	destination := filepath.Join(t.TempDir(), "artifact.tar.gz")
	checksumBody := fmt.Sprintf("%s  %s\n", sumHex, filepath.Base(destination))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/artifact":
			_, _ = w.Write(payload)
		case "/checksums.txt":
			_, _ = w.Write([]byte(checksumBody))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	err := DownloadFile(context.Background(), Options{
		URL:         server.URL + "/artifact",
		Destination: destination,
		ChecksumURL: server.URL + "/checksums.txt",
		NoProgress:  true,
		Retries:     1,
	})
	require.NoError(t, err)

	onDisk, err := os.ReadFile(destination)
	require.NoError(t, err)
	require.Equal(t, payload, onDisk)
}

// safeBuffer is a goroutine-safe bytes.Buffer for capturing concurrent writes
// from progress bar goroutines during tests.
type safeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *safeBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *safeBuffer) Bytes() []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	return bytes.Clone(b.buf.Bytes())
}

func TestDownloadProgressBarOutputEndsWithNewline(t *testing.T) {
	t.Parallel()

	payload := bytes.Repeat([]byte("x"), 1024)
	sum := sha256.Sum256(payload)
	sumHex := hex.EncodeToString(sum[:])

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(payload)))
		_, _ = w.Write(payload)
	}))
	defer server.Close()

	var progressBuf safeBuffer
	destination := filepath.Join(t.TempDir(), "artifact.bin")

	err := DownloadFile(context.Background(), Options{
		URL:            server.URL,
		Destination:    destination,
		ExpectedSHA256: sumHex,
		Retries:        1,
		ProgressWriter: &progressBuf,
	})
	require.NoError(t, err)

	output := progressBuf.Bytes()
	require.NotEmpty(t, output, "progress bar should have written output")
	require.True(t, bytes.HasSuffix(output, []byte("\n")),
		"download progress bar output must end with newline to prevent log overlap, got trailing bytes: %q",
		trailingBytes(output, 20))
}

func trailingBytes(b []byte, n int) []byte {
	if len(b) <= n {
		return b
	}
	return b[len(b)-n:]
}

func TestResolveExpectedChecksum(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa  model.bin\n"))
	}))
	defer server.Close()

	checksum, err := ResolveExpectedChecksum(context.Background(), server.URL, "model.bin", nil)
	require.NoError(t, err)
	require.Equal(t, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", checksum)
}
