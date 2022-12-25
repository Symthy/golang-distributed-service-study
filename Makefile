.PHONY: clean autogen build run list test e2e

clean:
	rm -f bin/*

autogen:
	protoc api/v1/*.proto --go_out=. --go_opt=paths=source_relative --proto_path=.

build:
	go build -o ./bin/main.exe cmd/server/main.go

run:
	go run cmd/server/main.go

list:
	go list ./...

test:
	go test `go list ./... | grep -v ./e2e`

e2e:
	go test -race ./e2e/...
