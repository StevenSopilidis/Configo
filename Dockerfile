FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" \
    -o collector .

RUN mkdir -p /data/config-store /data/snapshots

FROM gcr.io/distroless/base-debian12 AS runtime

WORKDIR /app

COPY --from=builder /app/collector /app/collector

EXPOSE 8080

ENTRYPOINT ["/app/collector"]
