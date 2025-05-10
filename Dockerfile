FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download 
RUN go mod tidy

COPY . .

RUN go build -o tag-nag

FROM debian:stable-slim

ARG TERRAFORM_VERSION

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    wget \
    unzip \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

RUN wget "https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip" && \
    unzip "terraform_${TERRAFORM_VERSION}_linux_amd64.zip" -d /usr/local/bin && \
    rm "terraform_${TERRAFORM_VERSION}_linux_amd64.zip" && \
    chmod +x /usr/local/bin/terraform

RUN groupadd -r appgroup && \
    useradd --no-log-init -r -g appgroup appuser

COPY --from=builder /app/tag-nag /usr/local/bin/tag-nag

USER appuser

WORKDIR /workspace

ENTRYPOINT ["tag-nag"]

CMD ["--help"]
