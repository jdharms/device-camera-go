.PHONY: build test clean prepare update

#GOOS=linux

GO=CGO_ENABLED=0 GO111MODULE=on go

MICROSERVICES=cmd/device-camera-go
.PHONY: $(MICROSERVICES)

VERSION=$(shell cat ./VERSION)

GOFLAGS=-ldflags "-X github.com/edgexfoundry/device-camera-go.Version=$(VERSION)"

build: $(MICROSERVICES)

cmd/device-camera-go:
	$(GO) build $(GOFLAGS) -o $@ ./cmd

test:
	go test -coverprofile=coverage.out ./...
	go vet ./...

lint: # requires the golint tool to be installed and available on your $PATH - (go get -u golang.org/x/lint/golint)
	go get -u golang.org/x/lint/golint
	golint ./... || true

coveragehtml:
	go tool cover -html=coverage.out -o coverage.html

format:
	gofmt -l .
	[ "`gofmt -l .`" = "" ]

update:
	$(GO) mod download

clean:
	rm -f $(MICROSERVICES)


