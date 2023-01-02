package log

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	writeData   = []byte("hello world")
	recordWidth = uint64(len(writeData)) + lenWidth
)

func TestStoreAppendRead(t *testing.T) {
	f, err := os.CreateTemp("", "store_append_read_test")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	s, err := newStore(f)
	require.NoError(t, err)

	testAppend(t, s)
	testRead(t, s)
	testReadAt(t, s)

	s, err = newStore(f)
	require.NoError(t, err)
	testRead(t, s)
}

func testAppend(t *testing.T, s *store) {
	t.Helper()
	for i := uint64(1); i < 4; i++ {
		n, pos, err := s.Append(writeData)
		require.NoError(t, err)
		fmt.Printf("temp: %v %v", pos+n, recordWidth*i)
		lines, _ := s.readAll()
		fmt.Printf("temp: %v", lines)
		require.Equal(t, pos+n, recordWidth*i)
	}
}

func testRead(t *testing.T, s *store) {
	t.Helper()
	var pos uint64 = 0
	for i := uint64(1); i < 4; i++ {
		readData, err := s.Read(pos)
		require.NoError(t, err)
		require.Equal(t, readData, writeData)
		pos += recordWidth
	}
}

func testReadAt(t *testing.T, s *store) {
	t.Helper()
	for i, offset := uint64(1), int64(0); i < 4; i++ {
		len := make([]byte, lenWidth)
		n, err := s.ReadAt(len, offset)
		require.NoError(t, err)
		require.Equal(t, lenWidth, n)

		offset += int64(n)
		size := enc.Uint64(len)
		b := make([]byte, size)

		n, err = s.ReadAt(b, offset)
		require.NoError(t, err)
		require.Equal(t, writeData, b)
		require.Equal(t, int(size), n)
		offset += int64(n)
	}
}

func TestStoreClose(t *testing.T) {
	f, err := os.CreateTemp("", "store_close_test")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	s, err := newStore(f)
	require.NoError(t, err)
	_, _, err = s.Append(writeData)
	require.NoError(t, err)

	beforeSize, err := s.getFileSize()
	require.NoError(t, err)

	err = s.Close()
	require.NoError(t, err)

	afterSize, err := s.getFileSize()
	require.NoError(t, err)
	require.True(t, afterSize > beforeSize)

}
