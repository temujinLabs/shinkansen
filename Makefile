BINARY=shinkansen
VERSION?=dev
GOFLAGS=-ldflags="-s -w -X main.version=$(VERSION)"

.PHONY: build test install clean lint docker

build:
	go build $(GOFLAGS) -o $(BINARY) ./cmd/shinkansen/

test:
	go test ./... -v

install: build
	cp $(BINARY) $(GOPATH)/bin/$(BINARY) 2>/dev/null || cp $(BINARY) /usr/local/bin/$(BINARY)

clean:
	rm -f $(BINARY)
	go clean

lint:
	go vet ./...

docker:
	docker build -t $(BINARY):$(VERSION) .

# Extract binary from Docker (cross-platform build)
docker-extract: docker
	docker create --name tmp-$(BINARY) $(BINARY):$(VERSION)
	docker cp tmp-$(BINARY):/usr/local/bin/$(BINARY) ./$(BINARY)
	docker rm tmp-$(BINARY)
