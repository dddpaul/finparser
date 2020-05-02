FROM golang:1.14.2 as builder
WORKDIR /go/src/github.com/dddpaul/finparser
ADD . ./
RUN make build-alpine

FROM alpine:latest
WORKDIR /app
COPY --from=builder /go/src/github.com/dddpaul/finparser/bin/finparser .

ENTRYPOINT ["./finparser"]
