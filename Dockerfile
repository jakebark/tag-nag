FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /tag-nag main.go

FROM gcr.io/distroless/static-debian11 AS final

COPY --from=builder /tag-nag /usr/local/bin/tag-nag

ENTRYPOINT ["/usr/local/bin/tag-nag"] 
