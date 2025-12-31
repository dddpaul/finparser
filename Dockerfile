FROM golang:1.25 AS builder
WORKDIR /go/src/github.com/dddpaul/finparser
COPY . .

# Build arguments for cross-compilation
ARG TARGETARCH
ARG TARGETOS

# Install dependencies and build for target architecture
RUN go mod download && \
    CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -ldflags="-w -s" -o ./bin/finparser ./finparser.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /go/src/github.com/dddpaul/finparser/bin/finparser .

ENTRYPOINT ["./finparser"]
