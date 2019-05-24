.PHONY: build test clean prepare update

#GOOS=linux

GO=CGO_ENABLED=0 go

MICROSERVICES=cmd/device-camera-bosch
.PHONY: $(MICROSERVICES)

VERSION=$(shell cat ./VERSION)

GOFLAGS=-ldflags "-X github.com/dell-iot/device-camera-bosch.Version=$(VERSION)"

build: $(MICROSERVICES)
	go build ./...

cmd/device-camera-bosch:
	$(GO) build $(GOFLAGS) -o $@ ./cmd

test:
	go test ./... -cover

clean:
	rm -f $(MICROSERVICES)
