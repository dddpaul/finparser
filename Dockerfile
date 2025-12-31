FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /go/src/github.com/dddpaul/finparser

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build arguments for cross-compilation
ARG TARGETARCH
ARG TARGETOS

# Build the application
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -o ./bin/finparser ./finparser.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /go/src/github.com/dddpaul/finparser/bin/finparser .

ENTRYPOINT ["./finparser"]
