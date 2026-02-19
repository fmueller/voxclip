//go:build integration

package download

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDownloadFileEndToEndWithFixtureServer(t *testing.T) {
	payload := []byte("integration-payload")
	sum := sha256.Sum256(payload)
	sumHex := hex.EncodeToString(sum[:])

	target := filepath.Join(t.TempDir(), "model.bin")
	checksums := fmt.Sprintf("%s  %s\n", sumHex, filepath.Base(target))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/model.bin":
			_, _ = w.Write(payload)
		case "/checksums.txt":
			_, _ = w.Write([]byte(checksums))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	err := DownloadFile(context.Background(), Options{
		URL:         server.URL + "/model.bin",
		Destination: target,
		ChecksumURL: server.URL + "/checksums.txt",
		NoProgress:  true,
	})
	require.NoError(t, err)

	onDisk, err := os.ReadFile(target)
	require.NoError(t, err)
	require.Equal(t, payload, onDisk)
}
