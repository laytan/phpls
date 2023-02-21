gotest:
	go test ./... -cover -v -race -shuffle=on -timeout=5s -tags what

gotestrich:
	richgo test ./... -cover -v -race -shuffle=on -timeout=5s -tags what

gobuild:
	go generate ./... && go build -o elephp cmd/elephp/main.go

gotestbuild:
	go generate ./... && go build -o elephp -tags what cmd/elephp/main.go

build-unsafenil:
	cd third_party/unsafenil && go build -o ../../unsafenil.so -buildmode=plugin plugin/unsafenil.go
