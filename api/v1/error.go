package log_v1

import (
	"fmt"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
)

type ErrOffsetOutOfRange struct {
	Offset uint64
}

func NewErrOffsetOutOfRange(offset uint64) ErrOffsetOutOfRange {
	return ErrOffsetOutOfRange{Offset: offset}
}

func (e *ErrOffsetOutOfRange) gRPCStatus() *status.Status {
	st := status.New(404, fmt.Sprintf("offset out of range: %d", e.Offset))
	d := &errdetails.LocalizedMessage{
		Locale:  "en-US",
		Message: fmt.Sprintf("The requeted offset is outside: %d", e.Offset),
	}
	std, err := st.WithDetails(d)
	if err != nil {
		return st
	}
	return std
}

func (e ErrOffsetOutOfRange) Error() string {
	return e.gRPCStatus().Err().Error()
}
