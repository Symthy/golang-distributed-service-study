package main

import (
	"log"

	"github.com/Symthy/golang-distributed-service-study/internal/http/server"
)

func main() {
	srv := server.NewHttpServer(":8080")
	log.Fatal(srv.ListenAndServe())
}
