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

testdeps:
	go get -v github.com/golang/lint/golint
	go get -v honnef.co/go/simple/cmd/gosimple
	go get -t -v ./...
