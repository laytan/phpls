test:
	go test ./... -cover -race -timeout=5s

build:
	go build -o elephp cmd/main.go
