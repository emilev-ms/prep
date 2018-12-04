all: test lint

generate:
	go run ./cmd/prep/prep.go ./cmd/prep/interface_finder.go -f github.com/Melsoft-Games/prep -t

lint:
	golint ./... && go vet ./...

test: generate
	go test -race ./...
