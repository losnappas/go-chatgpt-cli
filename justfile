build:
  CGO_ENABLED=0 go build ./cmd/chatgpt_cli/chatgpt_cli.go

test:
  go test ./internal/...
