test:
	go test ./... -cover -race -shuffle=on -timeout=5s

build:
	go build -o elephp cmd/main.go
