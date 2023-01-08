package log_v1

import (
	"fmt"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ErrOffsetOutOfRange struct {
	Offset uint64
}

// ポインタレシーバーにする必要はない
func (e ErrOffsetOutOfRange) GRPCStatus() *status.Status {
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

func (e ErrOffsetOutOfRange) Code() codes.Code {
	return status.Code(e.GRPCStatus().Err())
}

func (e ErrOffsetOutOfRange) Error() string {
	return e.GRPCStatus().Err().Error()
}
