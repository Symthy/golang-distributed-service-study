package log

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIndex(t *testing.T) {
	f, err := os.CreateTemp(os.TempDir(), "index_test")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	c := Config{}
	c.Segment.MaxIndexBytes = 1024

	idx, err := newIndex(f, c)
	require.NoError(t, err)

	_, _, err = idx.ReadLast()
	require.Error(t, err)
	require.Equal(t, f.Name(), idx.Name())

	entries := []struct {
		Off uint32
		Pos uint64
	}{
		{Off: 0, Pos: 0},
		{Off: 1, Pos: 10},
	}
	for _, tt := range entries {
		err = idx.Write(tt.Off, tt.Pos)
		require.NoError(t, err)
		off, pos, err := idx.Read(uint32(tt.Off))
		require.NoError(t, err)
		require.Equal(t, tt.Off, off)
		require.Equal(t, tt.Pos, pos)
	}

	t.Run("out of index", func(t *testing.T) {
		_, _, err := idx.Read(uint32(len(entries)))
		require.Equal(t, io.EOF, err)
	})
	err = idx.Close()
	require.NoError(t, err)

	t.Run("read last index", func(t *testing.T) {
		f, _ = os.OpenFile(f.Name(), os.O_RDWR, 0600)
		idx, err = newIndex(f, c)
		require.NoError(t, err)
		off, pos, err := idx.ReadLast()
		require.NoError(t, err)
		require.Equal(t, entries[1].Off, off)
		require.Equal(t, entries[1].Pos, pos)
	})

}
