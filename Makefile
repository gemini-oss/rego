# Makefile \
:Copyright: (c) 2024 by Gemini Space Station, LLC., see AUTHORS for more info \
:License: See the LICENSE file for details \
:Author: Anthony Dardano <anthony.dardano@gemini.com>

BINARY=rego

# Builds the project
build:
	go build -tags "$(tags)" -o ${BINARY} main.go

# Runs tests
test:
	go test -v ./...

# Generates markdown documentation for Hugo
docs:
	./gen_hugo_index.sh

# Starts the hugo test server
server:
	hugo server -s hugo --disableFastRender

# Formats the code
pretty:
	gofmt -s -w .

# Runs the application
run: build
	./${BINARY}

# Cleans the binary
clean:
	go clean
	rm ${BINARY}
