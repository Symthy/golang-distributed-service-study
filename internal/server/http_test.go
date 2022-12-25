package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ref: https://github.com/gorilla/mux#testing-handlers

var (
	record1 = LogRecord{Value: []byte("TGV0J3MgR28gIzEK"), Offset: 0}
	record2 = LogRecord{Value: []byte("TGV0J3MgR28gIzIK"), Offset: 1}
	record3 = LogRecord{Value: []byte("TGV0J3MgR28gIzMK"), Offset: 2}

	unknwonError = fmt.Errorf("dummy error")
)

func executeComsumeHandler(t *testing.T, srv *httpServer, input ConsumeRequest, rr *httptest.ResponseRecorder) {
	reqJson, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest("GET", "/", bytes.NewBuffer(reqJson))
	if err != nil {
		t.Fatal(err)
	}
	handler := http.HandlerFunc(srv.handleConsume)
	handler.ServeHTTP(rr, req)
}

func executeProduceHandler(t *testing.T, srv *httpServer, input ProduceRequest, rr *httptest.ResponseRecorder) {
	reqJson, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(reqJson))
	if err != nil {
		t.Fatal(err)
	}
	handler := http.HandlerFunc(srv.handleProduce)
	handler.ServeHTTP(rr, req)
}

func TestHandleConsumeSuccess(t *testing.T) {
	srv := newHttpServerWithLog(
		&Log{records: []LogRecord{record1, record2}},
	)
	cases := []struct {
		title            string
		input            ConsumeRequest
		expectedResponse ConsumeResponse
	}{
		{
			title:            "GET offset 0",
			input:            ConsumeRequest{Offset: uint64(0)},
			expectedResponse: ConsumeResponse{Record: record1},
		},
		{
			title:            "GET offset 1",
			input:            ConsumeRequest{Offset: uint64(1)},
			expectedResponse: ConsumeResponse{Record: record2},
		},
	}

	for _, tt := range cases {
		t.Run(tt.title, func(t *testing.T) {
			responseRecorder := httptest.NewRecorder()
			executeComsumeHandler(t, srv, tt.input, responseRecorder)
			assert.Equal(t, http.StatusOK, responseRecorder.Code)
			expected, err := json.Marshal(tt.expectedResponse)
			if err != nil {
				t.Error(err)
			}
			actual := strings.TrimRight(responseRecorder.Body.String(), "\n")
			assert.Equal(t, string(expected), actual)
		})
	}
}

type MockedLogForErrorCase struct {
}

func (*MockedLogForErrorCase) Append(record LogRecord) (uint64, error) {
	return 0, unknwonError
}
func (*MockedLogForErrorCase) Read(offset uint64) (LogRecord, error) {
	return LogRecord{}, unknwonError
}

func TestHandleConsumeFailure(t *testing.T) {
	cases := []struct {
		title          string
		inputSrv       *httpServer
		expectedStatus int
		expectedError  error
	}{
		{
			title: "Not Found Error",
			inputSrv: newHttpServerWithLog(
				&Log{records: []LogRecord{record1, record2}},
			),
			expectedStatus: http.StatusNotFound,
			expectedError:  ErrOffsetNotFound,
		},
		{
			title: "Internal Server Error",
			inputSrv: newHttpServerWithLog(
				&MockedLogForErrorCase{},
			),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  unknwonError,
		},
	}
	req := ConsumeRequest{Offset: uint64(3)}

	for _, tt := range cases {
		t.Run(tt.title, func(t *testing.T) {
			responseRecorder := httptest.NewRecorder()
			executeComsumeHandler(t, tt.inputSrv, req, responseRecorder)
			assert.Equal(t, tt.expectedStatus, responseRecorder.Code)

			actual := strings.TrimRight(responseRecorder.Body.String(), "\n")
			assert.Equal(t, tt.expectedError.Error(), actual)
		})
	}
}

func TestHandleProduceSuccess(t *testing.T) {
	srv := newHttpServerWithLog(
		&Log{records: []LogRecord{record1, record2}},
	)

	cases := []struct {
		title            string
		input            ProduceRequest
		expectedResponse ProduceResponse
	}{
		{
			title:            "POST 1",
			input:            ProduceRequest{Record: record3},
			expectedResponse: ProduceResponse{Offset: 2},
		},
	}

	for _, tt := range cases {
		t.Run(tt.title, func(t *testing.T) {
			responseRecorder := httptest.NewRecorder()
			executeProduceHandler(t, srv, tt.input, responseRecorder)

			assert.Equal(t, http.StatusOK, responseRecorder.Code)
			expected, err := json.Marshal(tt.expectedResponse)
			if err != nil {
				t.Error(err)
			}
			actual := strings.TrimRight(responseRecorder.Body.String(), "\n")
			assert.Equal(t, string(expected), actual)
		})
	}
}

func TestHandleProduceFailure(t *testing.T) {
	cases := []struct {
		title          string
		inputSrv       *httpServer
		expectedStatus int
		expectedError  error
	}{
		{
			title: "Internal Server Error",
			inputSrv: newHttpServerWithLog(
				&MockedLogForErrorCase{},
			),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  unknwonError,
		},
	}
	req := ProduceRequest{Record: record2}

	for _, tt := range cases {
		t.Run(tt.title, func(t *testing.T) {
			responseRecorder := httptest.NewRecorder()
			executeProduceHandler(t, tt.inputSrv, req, responseRecorder)
			assert.Equal(t, tt.expectedStatus, responseRecorder.Code)

			actual := strings.TrimRight(responseRecorder.Body.String(), "\n")
			assert.Equal(t, tt.expectedError.Error(), actual)
		})
	}
}
