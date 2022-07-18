# Elephp

## Commands

Example running, use `go run cmd/main.go -h` for help:
```bash
go run cmd/main.go --stdio
```

Run the LS with [statsviz](https://github.com/arl/statsviz) on [http://localhost:6060/debug/statsviz](http://localhost:6060/debug/statsviz):
```bash
go run cmd/main.go --statsviz
```


Run all tests with coverage:
```bash
go test ./... --cover -timeout=1s
```
