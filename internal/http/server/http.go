package server

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func NewHttpServer(addr string) *http.Server {
	httpsrv := newHttpServer()
	router := mux.NewRouter()
	router.HandleFunc("/", httpsrv.handleProduce).Methods("POST")
	router.HandleFunc("/", httpsrv.handleConsume).Methods("GET")
	return &http.Server{
		Addr:    addr,
		Handler: router,
	}
}

type httpServer struct {
	log Store
}

func newHttpServer() *httpServer {
	return &httpServer{
		log: NewLog(),
	}
}

func newHttpServerWithLog(log Store) *httpServer {
	return &httpServer{
		log: log,
	}
}

type ProduceRequest struct {
	Record LogRecord `json:"record"`
}

type ProduceResponse struct {
	Offset uint64 `json:"offset"`
}

type ConsumeRequest struct {
	Offset uint64 `json:"offset"`
}

type ConsumeResponse struct {
	Record LogRecord `json:"record"`
}

func (s *httpServer) handleProduce(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req ProduceRequest
	resolveRequest(w, r, &req)

	offset, err := s.log.Append(req.Record)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := ProduceResponse{Offset: offset}
	resolveResponse(w, r, &res)
}

func (s *httpServer) handleConsume(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req ConsumeRequest
	resolveRequest(w, r, &req)

	record, err := s.log.Read(req.Offset)
	if err == ErrOffsetNotFound {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := ConsumeResponse{Record: record}
	resolveResponse(w, r, &res)
}

func resolveRequest[T ProduceRequest | ConsumeRequest](w http.ResponseWriter, r *http.Request, req *T) {
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func resolveResponse[T ProduceResponse | ConsumeResponse](w http.ResponseWriter, r *http.Request, res *T) {
	w.Header().Add("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
