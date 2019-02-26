all: gotool
	@go build -v .
build-linux: 
	GOOS=linux GOARCH=amd64 go build -v -o bin .