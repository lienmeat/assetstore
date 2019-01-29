export GO111MODULE=on

all: deps build
install:
	go install cmd/server/main.go
build:
	go build cmd/server/main.go
test:
	go test -p 1 -v ./...
clean:
	go clean cmd/server/main.go
	rm -f main.go
deps:
	go build -v ./...
upgrade:
	go get -u