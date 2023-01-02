package log

import (
	"bufio"
	"encoding/binary"
	"io/fs"
	"io/ioutil"
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
	if err := binary.Write(s.buf, enc, uint64(len(p))); err != nil { // サイズの値書き込み？
		return 0, 0, err
	}
	w, err := s.buf.Write(p) // データ書き込み
	if err != nil {
		return 0, 0, err
	}

	w += lenWidth
	s.size += uint64(w)
	return uint64(w), pos, nil
}

func (s *store) Read(pos uint64) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.buf.Flush(); err != nil {
		return nil, err
	}
	size := make([]byte, lenWidth)
	if _, err := s.File.ReadAt(size, int64(pos)); err != nil { // サイズの値読み出し
		return nil, err
	}
	b := make([]byte, enc.Uint64(size))
	if _, err := s.File.ReadAt(b, int64(pos+lenWidth)); err != nil { // データ読み出し
		return nil, err
	}
	return b, nil
}

// io.ReadAt インターフェース実装
func (s *store) ReadAt(p []byte, offset int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.buf.Flush(); err != nil {
		return 0, err
	}
	return s.File.ReadAt(p, offset)
}

func (s *store) Flush() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.buf.Flush()
	if err != nil {
		return err
	}
	return nil
}

func (s *store) Close() error {
	if err := s.Flush(); err != nil {
		return err
	}
	return s.File.Close()
}

func (s *store) getFileSize() (size int64, err error) {
	f, err := os.OpenFile(s.File.Name(), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return 0, err
	}
	fi, err := f.Stat()
	if err != nil {
		return 0, err
	}
	return fi.Size(), nil
}

func (s *store) readAll() (lines string, err error) {
	fileContent, err := ioutil.ReadFile(s.File.Name())
	if err != nil {
		return "", err
	}
	return string(fileContent), nil
}
