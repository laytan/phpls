gotest:
	go test ./... -cover -v -race -shuffle=on -timeout=5s -tags what

gobuild:
	go build -o elephp cmd/main.go

gotestbuild:
	go build -o elephp -tags what cmd/main.go
