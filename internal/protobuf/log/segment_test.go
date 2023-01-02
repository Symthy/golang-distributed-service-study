package log

import (
	"io"
	"os"
	"testing"

	api "github.com/Symthy/golang-distributed-service-study/api/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestSegment(t *testing.T) {
	dir, _ := os.MkdirTemp("", "segment_test")
	defer os.RemoveAll(dir)

	expected := &api.Record{Value: []byte("hello world")}
	c := Config{}
	c.Segment.MaxStoreBytes = 1024
	c.Segment.MaxIndexBytes = entryWidth * 3

	s, err := newSegment(dir, 16, c)
	require.NoError(t, err)
	require.Equal(t, uint64(16), s.nextOffset)
	require.False(t, s.IsMaxed())

	t.Run("append and read test", func(t *testing.T) {
		for i := uint64(0); i < 3; i++ {
			off, err := s.Append(expected)
			require.NoError(t, err)
			require.Equal(t, 16+i, off)

			actual, err := s.Read(off)
			require.NoError(t, err)
			require.Equal(t, expected.Value, actual.Value)
		}

		_, err := s.Append(expected)
		require.Equal(t, io.EOF, err)
		require.True(t, s.IsMaxed())
	})

	require.NoError(t, s.Close())

	p, _ := proto.Marshal(expected)
	c.Segment.MaxStoreBytes = uint64(len(p)+lenWidth) * 4
	c.Segment.MaxIndexBytes = 1024
	t.Run("maxed store rebuild", func(t *testing.T) {
		s, err = newSegment(dir, 16, c)
		require.NoError(t, err)
		require.True(t, s.IsMaxed())
	})

	t.Run("remove and store rebuild", func(t *testing.T) {
		require.NoError(t, s.Remove())
		s, err = newSegment(dir, 16, c)
		require.NoError(t, err)
		require.False(t, s.IsMaxed())
		require.NoError(t, s.Close())
	})

}
