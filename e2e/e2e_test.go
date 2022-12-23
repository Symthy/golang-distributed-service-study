package e2e

import (
	"net/http"
	"testing"

	"github.com/Symthy/golang-distributed-service-study/internal/server"
	"github.com/gavv/httpexpect/v2"
)

func TestGetAndPostSuccess(t *testing.T) {
	e := httpexpect.New(t, "http://localhost:8080")
	val1 := "TGV0J3MgR28gIzEK"
	val2 := "TGV0J3MgR28gIzIK"
	val3 := "TGV0J3MgR28gIzMK"
	record1 := server.LogRecord{Value: []byte(val1), Offset: 0}
	record2 := server.LogRecord{Value: []byte(val2), Offset: 1}
	record3 := server.LogRecord{Value: []byte(val3), Offset: 2}

	t.Run("POST 1", func(t *testing.T) {
		r := e.POST("/").WithJSON(server.ProduceRequest{Record: record1}).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()
		r.Equal(server.ProduceResponse{Offset: 0})
	})
	t.Run("POST 2", func(t *testing.T) {
		r := e.POST("/").WithJSON(server.ProduceRequest{Record: record2}).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()
		r.Equal(server.ProduceResponse{Offset: 1})
	})
	t.Run("POST 2", func(t *testing.T) {
		r := e.POST("/").WithJSON(server.ProduceRequest{Record: record3}).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()
		r.Equal(server.ProduceResponse{Offset: 2})
	})

	t.Run("GET 1", func(t *testing.T) {
		r := e.GET("/").WithJSON(server.ConsumeRequest{Offset: 0}).
			Expect().              // レスポンスを取得
			Status(http.StatusOK). // assert: HTTP ステータスコードが OK
			JSON().                // assert: レスポンスボディが JSON
			Object()               // assert: JSON がオブジェクト
		r.Equal(server.ConsumeResponse{Record: record1})
	})
	t.Run("GET 2", func(t *testing.T) {
		r := e.GET("/").WithJSON(server.ConsumeRequest{Offset: 1}).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()
		r.Equal(server.ConsumeResponse{Record: record2})
	})
	t.Run("GET 3", func(t *testing.T) {
		r := e.GET("/").WithJSON(server.ConsumeRequest{Offset: 2}).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()
		r.Equal(server.ConsumeResponse{Record: record3})
	})
	t.Run("GET 3", func(t *testing.T) {
		r := e.GET("/").WithJSON(server.ConsumeRequest{Offset: 2}).
			Expect().
			Status(http.StatusOK).
			JSON().
			Object()
		r.Equal(server.ConsumeResponse{Record: record3})
	})
}
