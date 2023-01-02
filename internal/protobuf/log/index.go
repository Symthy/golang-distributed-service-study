package log

import (
	"fmt"
	"io"
	"os"

	"github.com/tysonmote/gommap"
)

const (
	offWidth uint64 = 4
	posWidth uint64 = 8
	entWidth        = offWidth + posWidth
)

type index struct {
	file *os.File
	mmap gommap.MMap
	size uint64
}

func newIndex(f *os.File, c Config) (*index, error) {
	idx := &index{file: f}
	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}
	idx.size = uint64(fi.Size())

	if err = os.Truncate(f.Name(), int64(c.Segment.MaxIndexBytes)); err != nil {
		return nil, err
	}

	if idx.mmap, err = gommap.Map(
		idx.file.Fd(),
		gommap.PROT_READ|gommap.PROT_WRITE,
		gommap.MAP_SHARED,
	); err != nil {
		return nil, err
	}
	return idx, nil
}

func (i *index) Name() string {
	return i.file.Name()
}

func (i *index) Write(off uint32, pos uint64) error {
	if i.isMaxed() {
		return io.EOF
	}
	enc.PutUint32(i.mmap[i.size:i.size+offWidth], off)
	enc.PutUint64(i.mmap[i.size+offWidth:i.size+entWidth], pos)
	i.size += uint64(entWidth)
	return nil
}

func (i *index) isMaxed() bool {
	return uint64(len(i.mmap)) < i.size+entWidth
}

func (i *index) Read(inOffset int64) (out uint32, pos uint64, err error) {
	if i.size == 0 {
		return 0, 0, io.EOF
	}
	offset := uint32(0)
	if inOffset == -1 {
		offset = uint32((i.size / entWidth) - 1) // 末尾取得
	} else {
		offset = uint32(inOffset)
	}
	fmt.Printf("offset: %v\n", offset)
	pos = uint64(offset) * entWidth
	if i.size < pos+entWidth {
		return 0, 0, io.EOF
	}

	out = enc.Uint32(i.mmap[pos : pos+offWidth])
	pos = enc.Uint64(i.mmap[pos+offWidth : pos+entWidth])
	return out, pos, nil
}

func (i *index) Close() error {
	if err := i.mmap.Sync(gommap.MS_ASYNC); err != nil {
		return fmt.Errorf("mmap sync error: %v", err)
	}
	if err := i.file.Sync(); err != nil {
		return fmt.Errorf("file sync error: %v", err)
	}
	// メモリ解放しないと file.Truncate() で以下エラーが発生する（Windows 特有の事象かもしれない…）
	// The requested operation cannot be performed on a file with a user-mapped section open.
	if err := i.mmap.UnsafeUnmap(); err != nil {
		return fmt.Errorf("mmmap unmap error: %v", err)
	}

	if err := i.file.Truncate(int64(i.size)); err != nil {
		return fmt.Errorf("file truncate error: %v", err)
	}
	return i.file.Close()
}
