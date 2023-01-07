package log

import (
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"

	api "github.com/Symthy/golang-distributed-service-study/api/v1"
	"github.com/Symthy/golang-distributed-service-study/internal/collections"
)

type Log struct {
	mu            sync.RWMutex
	dir           string
	conf          Config
	activeSegment *segment
	segments      []*segment
}

func NewLog(dir string, conf Config) (*Log, error) {
	if conf.Segment.MaxStoreBytes == 0 {
		conf.Segment.MaxStoreBytes = 1024
	}
	if conf.Segment.MaxIndexBytes == 0 {
		conf.Segment.MaxIndexBytes = 1024
	}
	l := &Log{
		dir:  dir,
		conf: conf,
	}
	return l, l.setup()
}

func (l *Log) setup() error {
	if err := l.restoreSegment(); err != nil {
		return err
	}

	if l.segments != nil {
		return nil
	}
	if err := l.newSegment(l.conf.Segment.InitialOffset); err != nil {
		return err
	}
	return nil
}

func (l *Log) newSegment(offset uint64) error {
	seg, err := newSegment(l.dir, offset, l.conf)
	if err != nil {
		return err
	}
	l.segments = append(l.segments, seg)
	l.activeSegment = seg
	return nil
}

func (l *Log) restoreSegment() error {
	files, err := os.ReadDir(l.dir)
	if err != nil {
		return err
	}

	var baseOffsets []uint64
	for _, file := range files {
		extention := path.Ext(file.Name()) // ファイル拡張子取得
		if extention != storeFileExtention {
			continue
		}
		offStr := strings.TrimSuffix(
			file.Name(),
			extention,
		)
		off, _ := strconv.ParseUint(offStr, 10, 0)
		baseOffsets = append(baseOffsets, off)
	}
	collections.SortAsc(baseOffsets)

	for i := 0; i < len(baseOffsets); i++ {
		if err = l.newSegment(baseOffsets[i]); err != nil {
			return err
		}
	}
	return nil
}

func (l *Log) Append(record *api.Record) (uint64, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.activeSegment.IsMaxed() {
		highestOffset := l.highestOffset()
		err := l.newSegment(highestOffset + 1)
		if err != nil {
			return 0, err
		}
	}

	off, err := l.activeSegment.Append(record)
	if err != nil {
		return 0, err
	}
	return off, nil
}

func (l *Log) highestOffset() uint64 {
	off := l.segments[len(l.segments)-1].nextOffset
	if off == 0 {
		return 0
	}
	return off - 1
}

func (l *Log) Read(offset uint64) (*api.Record, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	s := l.getSegmentIfContains(offset)
	if s == nil || s.nextOffset <= offset {
		return nil, api.NewErrOffsetOutOfRange(offset)
	}
	return s.Read(offset)
}

func (l *Log) getSegmentIfContains(offset uint64) *segment {
	for _, seg := range l.segments {
		if seg.baseOffset <= offset && offset < seg.nextOffset {
			return seg
		}
	}
	return nil
}

func (l *Log) Flush() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, segment := range l.segments {
		if err := segment.Flush(); err != nil {
			return err
		}
	}
	return nil
}

func (l *Log) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, segment := range l.segments {
		if err := segment.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (l *Log) Remove() error {
	if err := l.Close(); err != nil {
		return err
	}
	return os.RemoveAll(l.dir)
}

func (l *Log) Reset() error {
	if err := l.Remove(); err != nil {
		return err
	}
	return l.setup()
}

// 指定より小さいセグメントは削除。定期的に呼び出し不要になったものは削除
func (l *Log) Truncate(lowest uint64) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	var newSegments []*segment
	for _, s := range l.segments {
		if s.nextOffset > lowest+1 {
			newSegments = append(newSegments, s)
			continue
		}
		if err := s.Remove(); err != nil {
			return err
		}
	}
	l.segments = newSegments
	return nil
}

func (l *Log) Reader() io.Reader {
	l.mu.RLock()
	defer l.mu.RUnlock()
	readers := make([]io.Reader, len(l.segments))
	for i, segment := range l.segments {
		readers[i] = &originReader{segment.store, 0}
	}
	return io.MultiReader(readers...)
}

func (l *Log) LowestOffset() uint64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.segments[0].baseOffset
}

func (l *Log) HighestOffset() uint64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.highestOffset()
}

type originReader struct {
	*store
	offset int64
}

func (o *originReader) Read(p []byte) (int, error) {
	n, err := o.ReadAt(p, o.offset)
	o.offset = int64(n)
	return n, err
}
