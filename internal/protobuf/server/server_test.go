package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"

	api "github.com/Symthy/golang-distributed-service-study/api/v1"
	"github.com/Symthy/golang-distributed-service-study/internal/protobuf/auth"
	"github.com/Symthy/golang-distributed-service-study/internal/protobuf/config"
	"github.com/Symthy/golang-distributed-service-study/internal/protobuf/log"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

func setupTest(t *testing.T, fn func(*Config)) (
	rootClient, nobodyClient api.LogClient,
	cfg *Config,
	teardown func(),
) {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	newClient := func(crtPath, keyPath string) (
		*grpc.ClientConn,
		api.LogClient,
		[]grpc.DialOption,
	) {
		clientTlsConfig, err := config.SetupTlsConfig(config.TLSConfig{
			CertFile: crtPath,
			KeyFile:  keyPath,
			CAFile:   config.CAFile,
		})
		require.NoError(t, err)
		clientCreds := credentials.NewTLS(clientTlsConfig)
		clientOptions := []grpc.DialOption{
			grpc.WithTransportCredentials(clientCreds),
		}
		conn, err := grpc.Dial(
			l.Addr().String(),
			clientOptions...,
		)
		require.NoError(t, err)
		client := api.NewLogClient(conn)
		return conn, client, clientOptions
	}

	newServer := func(clog log.CommitLog) *grpc.Server {
		serverTLSConfig, err := config.SetupTlsConfig(config.TLSConfig{
			CertFile:      config.ServerCertFile,
			KeyFile:       config.ServerKeyFile,
			CAFile:        config.CAFile,
			ServerAddress: l.Addr().String(),
			Server:        true,
		})
		require.NoError(t, err)
		serverCreds := credentials.NewTLS(serverTLSConfig)
		authorizer, err := auth.New(config.ACLModelFile, config.ACLPolicyFile)
		require.NoError(t, err)
		cfg = &Config{
			CommitLog:  clog,
			Authorizer: authorizer,
		}
		if fn != nil {
			fn(cfg)
		}
		server, err := NewGRPCServer(cfg, grpc.Creds(serverCreds))
		require.NoError(t, err)
		return server
	}

	rootConn, rootClient, _ := newClient(
		config.RootClientCertFile,
		config.RootClientKeyFile,
	)
	nobodyConn, nobodyClient, _ := newClient(
		config.NobodyClientCertFile,
		config.NobodyClientKeyFile,
	)

	dir, err := os.MkdirTemp("", "server-test")
	require.NoError(t, err)
	clog, err := log.NewLog(dir, log.Config{})
	require.NoError(t, err)
	server := newServer(clog)

	go func() {
		server.Serve(l)
	}()

	teardown = func() {
		rootConn.Close()
		nobodyConn.Close()
		server.Stop()
		l.Close()
		clog.Remove()
	}
	return rootClient, nobodyClient, cfg, teardown
}

func TestServer(t *testing.T) {
	for senario, fn := range map[string]func(
		t *testing.T,
		rootClient api.LogClient,
		nobodyClient api.LogClient,
		config *Config,
	){
		"produce/consume message":         testProduceConsume,
		"produce/consume stream":          testProduceConsumeStream,
		"consume past log boundary fails": testConsumePastBoundary,
		"unauthorized fails":              testUnauthorized,
	} {
		t.Run(senario, func(t *testing.T) {
			rootClient, nobodyClient, config, teardown := setupTest(t, nil)
			defer teardown()
			fn(t, rootClient, nobodyClient, config)
		})
	}
}

func testProduceConsume(t *testing.T, client, _ api.LogClient, config *Config) {
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

func testConsumePastBoundary(t *testing.T, client, _ api.LogClient, config *Config) {
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

func testProduceConsumeStream(t *testing.T, client, _ api.LogClient, config *Config) {
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

func testUnauthorized(t *testing.T, _, nobodyClient api.LogClient, config *Config) {
	ctx := context.Background()
	produce, err := nobodyClient.Produce(ctx,
		&api.ProduceRequest{
			Record: &api.Record{
				Value: []byte("hello world"),
			},
		},
	)
	require.NotNil(t, produce)
	require.Equal(t, codes.PermissionDenied, status.Code(err))

	consume, err := nobodyClient.Consume(ctx,
		&api.ConsumeRequest{
			Offset: 0,
		},
	)
	require.NotNil(t, consume)
	require.Equal(t, codes.PermissionDenied, status.Code(err))
}
