VERSION := $(shell git describe --tags --always HEAD)
GOFLAGS = -ldflags "-X main.Version=$(VERSION)"

default: test

lint:
	golangci-lint run

build:
	go build $(GOFLAGS) -o aws-ecr-registry-cleaner ./cmd

test: build
	go test ./...

clean:
	rm -f aws-ecr-registry-cleaner

container:
	podman build -t aws-ecr-registry-cleaner .

update-deps:
	go get -t -u ./...
	go mod tidy
