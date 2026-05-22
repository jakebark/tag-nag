FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
RUN go mod tidy

COPY . .

RUN go build -o tag-nag

FROM debian:stable-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    unzip && \
    rm -rf /var/lib/apt/lists/*

RUN groupadd -r appgroup && \
    useradd --no-log-init -r -g appgroup appuser

COPY --from=builder /app/tag-nag /usr/local/bin/tag-nag

USER appuser

WORKDIR /workspace

ENTRYPOINT ["tag-nag"]

CMD ["--help"]
