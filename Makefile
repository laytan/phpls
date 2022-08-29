test:
	go test ./... -cover -race -shuffle=on -timeout=5s -v -tags what

build:
	go build -o elephp cmd/main.go
