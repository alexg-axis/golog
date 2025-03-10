package golog

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPool(t *testing.T) {
	buf := _bufferPool.Get()
	require.NotNil(t, buf)
	// Write 768 bytes (the max before the buffer's capacity exceeds )
	for i := 0; i < maxBufferSize; i++ {
		buf.WriteByte('C')
	}
	require.True(t, returnBuffer(buf))
	for i := 0; i <= maxBufferSize; i++ {
		buf.WriteByte('C')
	}
	require.False(t, returnBuffer(buf))
}
