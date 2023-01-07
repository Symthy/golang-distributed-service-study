package log

import (
	"fmt"
	"io"
	"os"
	"testing"

	api "github.com/Symthy/golang-distributed-service-study/api/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestLog(t *testing.T) {
	for senario, fn := range map[string]func(*testing.T, *Log){
		"append and read":                testAppendRead,
		"offset out of range error":      testOutOfRangeErr,
		"new log with existing segments": testNewExisting,
		"reader":                         testReader,
		"truncate":                       testTruncate,
	} {
		t.Run(senario, func(t *testing.T) {
			dir, err := os.MkdirTemp("", "log_test")
			require.NoError(t, err)
			defer os.RemoveAll(dir)

			conf := Config{}
			conf.Segment.MaxStoreBytes = 32
			log, err := NewLog(dir, conf)
			require.NoError(t, err)

			fn(t, log)
			require.NoError(t, log.Close())
		})
	}
}

func testAppendRead(t *testing.T, log *Log) {
	append := &api.Record{
		Value: []byte("Hello World"),
	}

	t.Run("append", func(t *testing.T) {
		offset, err := log.Append(append)
		require.NoError(t, err)
		require.Equal(t, uint64(0), offset)
	})

	t.Run("read", func(t *testing.T) {
		readdata, err := log.Read(uint64(0))
		require.NoError(t, err)
		require.Equal(t, append.Value, readdata.Value)
	})
}

func testOutOfRangeErr(t *testing.T, log *Log) {
	readdata, err := log.Read(1)
	require.Nil(t, readdata)
	require.ErrorAs(t, err, &api.ErrOffsetOutOfRange{})
	apiErr := err.(api.ErrOffsetOutOfRange)
	fmt.Printf("apiErr: %v", apiErr)
	require.Equal(t, uint64(1), apiErr.Offset)
}

func testNewExisting(t *testing.T, log *Log) {
	append := &api.Record{
		Value: []byte("Hello World"),
	}
	for i := 0; i < 3; i++ {
		_, err := log.Append(append)
		require.NoError(t, err)
	}
	require.NoError(t, log.Flush())
	assert.Equal(t, uint64(0), log.LowestOffset())
	assert.Equal(t, uint64(2), log.HighestOffset())

	newLog, err := NewLog(log.dir, log.conf)
	require.NoError(t, err)
	assert.Equal(t, uint64(0), newLog.LowestOffset())
	assert.Equal(t, uint64(2), newLog.HighestOffset())
	require.NoError(t, newLog.Close())
}

func testReader(t *testing.T, log *Log) {
	append := &api.Record{
		Value: []byte("Hello World"),
	}
	offset, err := log.Append(append)
	require.NoError(t, err)
	require.Equal(t, uint64(0), offset)

	reader := log.Reader()
	b, err := io.ReadAll(reader)
	require.NoError(t, err)

	readdata := &api.Record{}
	err = proto.Unmarshal(b[lenWidth:], readdata)
	require.NoError(t, err)
	require.Equal(t, append.Value, readdata.Value)
}

func testTruncate(t *testing.T, log *Log) {
	append := &api.Record{
		Value: []byte("Hello World"),
	}
	for i := 0; i < 3; i++ {
		_, err := log.Append(append)
		require.NoError(t, err)
	}

	err := log.Truncate(1)
	require.NoError(t, err)
	_, err = log.Read(0)
	require.Error(t, err)
}
