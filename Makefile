.PHONY: test build clean
DIRNAME=$(shell basename ${PWD})

test:
	go fmt
	gocyclo -over 12 $(shell find . -iname '*.go' -type f | grep -v /vendor/)
	golint -set_exit_status
	go test -v

build: test
	@mkdir -p build
	GOOS=linux GOARCH=amd64 go build -o build/${DIRNAME}_linux_amd64 main.go
	GOOS=darwin GOARCH=amd64 go build -o build/${DIRNAME}_darwin_amd64 main.go

clean:
	rm -rfv build
