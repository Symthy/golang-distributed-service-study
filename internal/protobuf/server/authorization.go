package server

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type Authorizer interface {
	Authorize(subject, object, action string) error
}

type subjectContextKey struct{}

// クライアントの証明書からサブジェクトを読み取ってRPCのコンテキストに書き込むinterceptor
func authenticate(ctx context.Context) (context.Context, error) {
	peer, ok := peer.FromContext(ctx) // 接続元の情報
	if !ok {
		return ctx, status.New(
			codes.Unknown,
			"could not find peer info",
		).Err()
	}

	if peer.AuthInfo == nil {
		return context.WithValue(ctx, subjectContextKey{}, ""), nil
	}

	tlsInfo := peer.AuthInfo.(credentials.TLSInfo)
	subject := tlsInfo.State.VerifiedChains[0][0].Subject.CommonName
	return context.WithValue(ctx, subjectContextKey{}, subject), nil
}

// クライアントの証明書のサブジェクトを返す
func subject(ctx context.Context) string {
	return ctx.Value(subjectContextKey{}).(string)
}
