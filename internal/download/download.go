package download

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
	"go.uber.org/zap"
	"golang.org/x/term"
)

var checksumPattern = regexp.MustCompile(`(?i)\b([a-f0-9]{64})\b`)

type Options struct {
	URL            string
	Destination    string
	ExpectedSHA256 string
	ChecksumURL    string
	Retries        int
	NoProgress     bool
	HTTPClient     *http.Client
	Logger         *zap.Logger
}

func DownloadFile(ctx context.Context, opts Options) error {
	if opts.URL == "" {
		return errors.New("download URL is required")
	}
	if opts.Destination == "" {
		return errors.New("destination path is required")
	}

	if opts.Retries <= 0 {
		opts.Retries = 3
	}

	if opts.HTTPClient == nil {
		opts.HTTPClient = &http.Client{Timeout: 10 * time.Minute}
	}

	if opts.Logger == nil {
		opts.Logger = zap.NewNop()
	}

	expected := strings.ToLower(strings.TrimSpace(opts.ExpectedSHA256))
	if expected == "" && opts.ChecksumURL != "" {
		resolved, err := ResolveExpectedChecksum(ctx, opts.ChecksumURL, filepath.Base(opts.Destination), opts.HTTPClient)
		if err != nil {
			return fmt.Errorf("fetch checksum: %w", err)
		}
		expected = resolved
	}

	if err := os.MkdirAll(filepath.Dir(opts.Destination), 0o755); err != nil {
		return fmt.Errorf("create destination directory: %w", err)
	}

	var lastErr error
	for attempt := 1; attempt <= opts.Retries; attempt++ {
		if attempt > 1 {
			opts.Logger.Warn("retrying download", zap.Int("attempt", attempt), zap.Int("max", opts.Retries), zap.String("url", opts.URL))
			time.Sleep(time.Duration(attempt) * 300 * time.Millisecond)
		}

		lastErr = downloadOnce(ctx, opts, expected)
		if lastErr == nil {
			return nil
		}
	}

	return lastErr
}

func ResolveExpectedChecksum(ctx context.Context, checksumURL, fileName string, client *http.Client) (string, error) {
	if strings.TrimSpace(checksumURL) == "" {
		return "", errors.New("checksum URL is required")
	}

	if client == nil {
		client = &http.Client{Timeout: 2 * time.Minute}
	}

	resolved, err := fetchExpectedChecksum(ctx, client, checksumURL, fileName)
	if err != nil {
		return "", err
	}

	return strings.ToLower(strings.TrimSpace(resolved)), nil
}

func ParseChecksum(content []byte, fileName string) (string, error) {
	lines := strings.Split(string(content), "\n")

	if fileName != "" {
		for _, line := range lines {
			if !strings.Contains(line, fileName) {
				continue
			}
			if checksum := parseChecksumFromLine(line); checksum != "" {
				return checksum, nil
			}
		}
	}

	for _, line := range lines {
		if checksum := parseChecksumFromLine(line); checksum != "" {
			return checksum, nil
		}
	}

	return "", errors.New("sha256 checksum not found")
}

func VerifyFileChecksum(path, expectedSHA256 string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open file for checksum: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("hash file: %w", err)
	}

	actual := hex.EncodeToString(h.Sum(nil))
	expected := strings.ToLower(strings.TrimSpace(expectedSHA256))
	if expected == "" {
		return nil
	}

	if actual != expected {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expected, actual)
	}

	return nil
}

func fetchExpectedChecksum(ctx context.Context, client *http.Client, checksumURL, fileName string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, checksumURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return ParseChecksum(content, fileName)
}

func parseChecksumFromLine(line string) string {
	match := checksumPattern.FindStringSubmatch(line)
	if len(match) < 2 {
		return ""
	}
	return strings.ToLower(match[1])
}

func downloadOnce(ctx context.Context, opts Options, expectedChecksum string) error {
	tempPath := opts.Destination + ".part"
	_ = os.Remove(tempPath)

	outFile, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	success := false
	defer func() {
		_ = outFile.Close()
		if !success {
			_ = os.Remove(tempPath)
		}
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, opts.URL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "voxclip/1")

	resp, err := opts.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("download request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	hash := sha256.New()
	writer := io.MultiWriter(outFile, hash)

	var bar *progressbar.ProgressBar
	if shouldRenderProgress(opts.NoProgress, resp.ContentLength) {
		bar = progressbar.NewOptions64(
			resp.ContentLength,
			progressbar.OptionSetDescription("downloading"),
			progressbar.OptionSetWidth(20),
			progressbar.OptionShowBytes(true),
			progressbar.OptionThrottle(65*time.Millisecond),
			progressbar.OptionSetRenderBlankState(true),
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionClearOnFinish(),
		)
		writer = io.MultiWriter(outFile, hash, bar)
	}

	if _, err := io.Copy(writer, resp.Body); err != nil {
		return fmt.Errorf("download body: %w", err)
	}

	if bar != nil {
		_ = bar.Finish()
	}

	if err := outFile.Sync(); err != nil {
		return fmt.Errorf("sync temp file: %w", err)
	}

	actualChecksum := hex.EncodeToString(hash.Sum(nil))
	if expectedChecksum != "" && actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	if err := outFile.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	if err := os.Rename(tempPath, opts.Destination); err != nil {
		return fmt.Errorf("move temp file into destination: %w", err)
	}

	success = true
	return nil
}

func shouldRenderProgress(noProgress bool, contentLength int64) bool {
	if noProgress {
		return false
	}
	if contentLength <= 0 {
		return false
	}
	return term.IsTerminal(int(os.Stderr.Fd()))
}
