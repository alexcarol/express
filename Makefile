check: linters test

linters: fmt lint vet simple

test:
	go test -race -v ./...

fmt:
	go fmt ./...

simple:
	gosimple ./...

lint:
	golint ./...

vet:
	go vet ./...
