run-server:
	@go run cmd/server/main.go

run-worker:
	@go run cmd/worker/main.go

run-k8s-worker:
	@go run cmd/k8s-worker/main.go

sqlc:
	@go run github.com/sqlc-dev/sqlc/cmd/sqlc@latest generate

gqlgen:
	@go run github.com/99designs/gqlgen@latest generate

test:
	@go test ./...
