package log

import (
	"bufio"
	"encoding/binary"
	"io/fs"
	"os"
	"sync"
)

var (
	enc = binary.BigEndian
)

const (
	lenWidth = 8
)

type store struct {
	*os.File
	mu   sync.Mutex
	buf  *bufio.Writer
	size uint64
}

func newStore(f *os.File) (*store, error) {
	fileInfo, err := getFileInfoIfExists(f.Name())
	if err != nil {
		return nil, err
	}
	return &store{
		File: f,
		size: uint64(fileInfo.Size()),
		buf:  bufio.NewWriter(f),
	}, nil
}

func getFileInfoIfExists(fileName string) (fs.FileInfo, error) {
	return os.Stat(fileName)
}

func (s *store) Append(p []byte) (num uint64, pos uint64, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pos = s.size
	if err := binary.Write(s.buf, enc, uint64(len(p))); err != nil {
		return 0, 0, err
	}
	w, err := s.buf.Write(p)
	if err != nil {
		return 0, 0, err
	}

	s.size += uint64(w + lenWidth)
	return uint64(w), pos, nil
}
