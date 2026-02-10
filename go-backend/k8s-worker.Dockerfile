# syntax=docker/dockerfile:1
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o k8s-worker cmd/k8s-worker/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata git

WORKDIR /app

COPY --from=builder /app/k8s-worker .
COPY --from=builder /app/application.yaml .

CMD ["./k8s-worker"]

