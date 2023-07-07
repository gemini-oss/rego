BINARY=rego

# Builds the project
build:
	go build -tags "$(tags)" -o ${BINARY} main.go

# Runs tests
test:
	go test -v ./...

# Generates documentation
doc:
	~/go/bin/gomarkdoc ./... --output 'docs/{{.Dir}}/README.md' --exclude-dirs ./pkg/internal/tests/...

pretty:
	gofmt -s -w .

# Runs the application
run: build
	./${BINARY}

# Cleans the binary
clean:
	go clean
	rm ${BINARY}
