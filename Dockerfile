FROM golang:1.22-alpine AS builder

WORKDIR /build

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=${VERSION}" -o shinkansen ./cmd/shinkansen/

# Minimal runtime
FROM alpine:3.19

RUN apk add --no-cache ca-certificates

COPY --from=builder /build/shinkansen /usr/local/bin/shinkansen

ENTRYPOINT ["shinkansen"]
