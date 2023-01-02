.PHONY: clean autogen lint build run list test testv e2e

clean:
	rm -f bin/*

autogen:
	protoc api/v1/*.proto \
		--go_out=. \
		--go-grpc_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_opt=paths=source_relative \
		--proto_path=.

lint:
	gofmt -l -s -w .

build:
	go build -o ./bin/main.exe cmd/server/main.go

run:
	go run cmd/server/main.go

list:
	go list ./...

test:
	go test `go list ./... | grep -v ./e2e`

testv:
	go test -v `go list ./... | grep -v ./e2e`

e2e:
	go test -race ./e2e/...
