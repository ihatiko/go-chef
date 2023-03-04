FROM golang:alpine AS builder
RUN apk add git

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build

COPY go.sum .
COPY go.mod .
RUN go mod download

COPY . .

RUN go build -o main .

WORKDIR /dist

RUN cp /build/main .

FROM alpine

COPY --from=builder /dist/main /

COPY config /config

ENTRYPOINT ["/main"]