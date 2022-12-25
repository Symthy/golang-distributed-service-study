.PHONY: clean build run list test e2e

clean:
	rm -f bin/*

build:
	go build -o ./bin/main.exe cmd/server/main.go

run:
	go run cmd/server/main.go

list:
	go list ./...

test:
	go test `go list ./... | grep -v ./e2e`

e2e:
	go test ./e2e/...
