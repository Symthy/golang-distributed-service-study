package e2e

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"testing"

	"github.com/Symthy/golang-distributed-service-study/internal/http/server"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/suite"
)

type e2eTestSuite struct {
	suite.Suite
	srv *http.Server
}

func (s *e2eTestSuite) SetupSuite() {
	s.srv = server.NewHttpServer(":8080")
}

func (s *e2eTestSuite) SetupTest() {
	// 動的にポートを選択するので並行テストが可能。
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		fmt.Println("server run")
		if err := s.srv.Serve(l); err != http.ErrServerClosed {
			s.T().Fatalf("HTTP server ListenAndServe: %v", err)
		}
		// サーバが終了したことを通知。
		close(idleConnsClosed)
	}()
}

func (s *e2eTestSuite) TearDownSuite() {
	fmt.Println("server shutdown")
	if err := s.srv.Shutdown(context.Background()); err != nil {
		s.T().Fatalf("HTTP server Shutdown: %v", err)
	}
}

func (s *e2eTestSuite) TestGetAndPostSuccess() {

	e := httpexpect.New(s.T(), "http://localhost:8080")
	val1 := "TGV0J3MgR28gIzEK"
	val2 := "TGV0J3MgR28gIzIK"
	val3 := "TGV0J3MgR28gIzMK"
	record1 := server.LogRecord{Value: []byte(val1), Offset: 0}
	record2 := server.LogRecord{Value: []byte(val2), Offset: 1}
	record3 := server.LogRecord{Value: []byte(val3), Offset: 2}

	s.T().Run("POST 1", func(t *testing.T) {
		r := e.POST("/").WithJSON(server.ProduceRequest{Record: record1}).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()
		r.Equal(server.ProduceResponse{Offset: 0})
	})
	s.T().Run("POST 2", func(t *testing.T) {
		r := e.POST("/").WithJSON(server.ProduceRequest{Record: record2}).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()
		r.Equal(server.ProduceResponse{Offset: 1})
	})
	s.T().Run("POST 2", func(t *testing.T) {
		r := e.POST("/").WithJSON(server.ProduceRequest{Record: record3}).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()
		r.Equal(server.ProduceResponse{Offset: 2})
	})

	s.T().Run("GET 1", func(t *testing.T) {
		r := e.GET("/").WithJSON(server.ConsumeRequest{Offset: 0}).
			Expect().              // レスポンスを取得
			Status(http.StatusOK). // assert: HTTP ステータスコードが OK
			JSON().                // assert: レスポンスボディが JSON
			Object()               // assert: JSON がオブジェクト
		r.Equal(server.ConsumeResponse{Record: record1})
	})
	s.T().Run("GET 2", func(t *testing.T) {
		r := e.GET("/").WithJSON(server.ConsumeRequest{Offset: 1}).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()
		r.Equal(server.ConsumeResponse{Record: record2})
	})
	s.T().Run("GET 3", func(t *testing.T) {
		r := e.GET("/").WithJSON(server.ConsumeRequest{Offset: 2}).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()
		r.Equal(server.ConsumeResponse{Record: record3})
	})
}

func TestE2ETestSuite(t *testing.T) {
	suite.Run(t, new(e2eTestSuite))
}
