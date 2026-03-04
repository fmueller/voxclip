package testutil

import (
	"bytes"
	"sync"
)

// SafeBuffer is a goroutine-safe bytes.Buffer for capturing concurrent writes
// from progress bar goroutines during tests.
type SafeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *SafeBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *SafeBuffer) Bytes() []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	return bytes.Clone(b.buf.Bytes())
}

// TrailingBytes returns the last n bytes of b, or all of b if len(b) <= n.
func TrailingBytes(b []byte, n int) []byte {
	if len(b) <= n {
		return b
	}
	return b[len(b)-n:]
}
