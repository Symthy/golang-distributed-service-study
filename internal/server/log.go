package server

import (
	"fmt"
	"sync"
)

var ErrOffsetNotFound = fmt.Errorf("offset not found")

type Store interface {
	Append(record LogRecord) (uint64, error)
	Read(offset uint64) (LogRecord, error)
}

var _ Store = (*Log)(nil)

type LogRecord struct {
	Value  []byte `json:"value"`
	Offset uint64 `json:"offset"`
}

type Log struct {
	mutex   sync.Mutex
	records []LogRecord
}

func NewLog() *Log {
	return &Log{}
}

func (c *Log) Append(record LogRecord) (uint64, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	record.Offset = uint64(len(c.records))
	c.records = append(c.records, record)
	return record.Offset, nil
}

func (c *Log) Read(offset uint64) (LogRecord, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if offset >= uint64(len(c.records)) {
		return LogRecord{}, ErrOffsetNotFound
	}
	return c.records[offset], nil
}
