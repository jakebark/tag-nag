FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

RUN go build -o tag-nag

FROM debian:stable-slim

WORKDIR /app

COPY --from=builder /app/tag-nag /usr/local/bin/tag-nag

ENTRYPOINT ["tag-nag"]

