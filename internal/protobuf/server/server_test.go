package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"

	api "github.com/Symthy/golang-distributed-service-study/api/v1"
	"github.com/Symthy/golang-distributed-service-study/internal/protobuf/config"
	"github.com/Symthy/golang-distributed-service-study/internal/protobuf/log"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

func setupTest(t *testing.T, fn func(*Config)) (client api.LogClient, cfg *Config, teardown func()) {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	clientTlsConfig, err := config.SetupTlsConfig(config.TLSConfig{
		CertFile: config.ClientCertFile,
		KeyFile:  config.ClientKeyFile,
		CAFile:   config.CAFile,
	})
	require.NoError(t, err)
	clientCreds := credentials.NewTLS(clientTlsConfig)
	// clientOption := []grpc.DialOption{
	// 	grpc.WithTransportCredentials(insecure.NewCredentials()),
	// }
	cc, err := grpc.Dial(
		l.Addr().String(),
		// clientOption...
		clientCreds,
	)
	require.NoError(t, err)
	client = api.NewLogClient(cc)

	serverTLSConfig, err := config.SetupTlsConfig(config.TLSConfig{
		CertFile:      config.ServerCertFile,
		KeyFile:       config.ServerKeyFile,
		CAFile:        config.CAFile,
		ServerAddress: l.Addr().String(),
		Server:        true,
	})
	require.NoError(t, err)
	serverCreds := credentials.NewTLS(serverTLSConfig)
	dir, err := os.MkdirTemp("", "server-test")
	require.NoError(t, err)
	clog, err := log.NewLog(dir, log.Config{})
	require.NoError(t, err)
	cfg = &Config{
		CommitLog: clog,
	}
	if fn != nil {
		fn(cfg)
	}
	server, err := NewGRPCServer(cfg, grpc.Creds(serverCreds))
	require.NoError(t, err)

	go func() {
		server.Serve(l)
	}()

	teardown = func() {
		cc.Close()
		server.Stop()
		l.Close()
		clog.Remove()
	}
	return client, cfg, teardown
}

func TestServer(t *testing.T) {
	for senario, fn := range map[string]func(
		t *testing.T,
		client api.LogClient,
		config *Config,
	){
		"produce/consume message":         testProduceConsume,
		"produce/consume stream":          testProduceConsumeStream,
		"consume past log boundary fails": testConsumePastBoundary,
	} {
		t.Run(senario, func(t *testing.T) {
			client, config, teardown := setupTest(t, nil)
			defer teardown()
			fn(t, client, config)
		})
	}
}

func testProduceConsume(t *testing.T, client api.LogClient, config *Config) {
	ctx := context.Background()

	want := &api.Record{
		Value: []byte("hello world"),
	}
	produce, err := client.Produce(
		ctx,
		&api.ProduceRequest{
			Record: want,
		},
	)
	require.NoError(t, err)
	want.Offset = produce.Offset

	consume, err := client.Consume(ctx, &api.ConsumeRequest{
		Offset: want.Offset,
	})
	require.NoError(t, err)
	require.Equal(t, want.Offset, consume.Record.Offset)
	require.Equal(t, want.Value, consume.Record.Value)
}

func testConsumePastBoundary(t *testing.T, client api.LogClient, config *Config) {
	ctx := context.Background()

	produce, err := client.Produce(
		ctx,
		&api.ProduceRequest{
			Record: &api.Record{Value: []byte("hello world")},
		},
	)
	require.NoError(t, err)

	consume, err := client.Consume(ctx, &api.ConsumeRequest{
		Offset: produce.Offset + 1,
	})
	require.Nil(t, consume)
	require.Error(t, err)
	fmt.Printf("error: %v\n", err)
	got := status.Code(err)
	require.Equal(t, api.ErrOffsetOutOfRange{}.Code().String(), got.String())
}

func testProduceConsumeStream(t *testing.T, client api.LogClient, config *Config) {
	ctx := context.Background()

	records := []*api.Record{
		{
			Value:  []byte("first message"),
			Offset: 0,
		},
		{
			Value:  []byte("second message"),
			Offset: 1,
		},
		{
			Value:  []byte("third message"),
			Offset: 2,
		},
	}

	t.Run("produce stream", func(t *testing.T) {
		stream, err := client.ProduceStream(ctx)
		require.NoError(t, err)
		for _, record := range records {
			err = stream.Send(&api.ProduceRequest{
				Record: record,
			})
			require.NoError(t, err)

			res, err := stream.Recv()
			require.NoError(t, err)
			require.Equal(t, record.Offset, res.Offset)
		}
	})

	t.Run("consumer stream", func(t *testing.T) {
		stream, err := client.ConsumeStream(
			ctx,
			&api.ConsumeRequest{Offset: 1},
		)
		require.NoError(t, err)
		for _, record := range records[1:] {
			res, err := stream.Recv()
			require.NoError(t, err)
			require.Equal(t, record.Offset, res.Record.Offset)
			require.Equal(t, record.Value, res.Record.Value)
		}
	})
}
