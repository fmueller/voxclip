package testutil

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSafeBufferConcurrentWrites(t *testing.T) {
	t.Parallel()

	var buf SafeBuffer
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = buf.Write([]byte("x"))
		}()
	}
	wg.Wait()
	require.Len(t, buf.Bytes(), 10)
}

func TestTrailingBytesShort(t *testing.T) {
	t.Parallel()
	require.Equal(t, []byte("abc"), TrailingBytes([]byte("abc"), 5))
}

func TestTrailingBytesLong(t *testing.T) {
	t.Parallel()
	require.Equal(t, []byte("cde"), TrailingBytes([]byte("abcde"), 3))
}
