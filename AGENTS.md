# AGENTS.md

## Repository

Go CLI health-check probe tool. Pure stdlib, no external dependencies.

## Build & Run

- `go build -o healthcheck .`
- `go run main.go -target http://localhost:8080/health -target http://localhost:9090/health -interval 10s -timeout 5s`
- `go test ./...` (no tests yet)
- `go vet ./...` for static analysis

## Conventions

- Default branch: `main`
- Single-file architecture (`main.go`), no packages to split yet
- `-target` flag is repeatable and required
