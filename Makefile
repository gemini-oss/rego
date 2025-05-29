# Makefile \
:Copyright: (c) 2024 by Gemini Space Station, LLC., see AUTHORS for more info \
:License: See the LICENSE file for details \
:Author: Anthony Dardano <anthony.dardano@gemini.com>

BINARY=rego

# Builds the project
build:
	go build -tags "$(tags)" -o ${BINARY} main.go

# Builds the project for Windows
windows:
	GOOS=windows GOARCH=amd64 go build -tags "$(tags)" -o ${BINARY} main.go

# Runs tests
test:
	go test -v ./...

# Generates markdown documentation for Hugo
docs:
	./gen_hugo_index.sh

# Starts the hugo test server
server:
	git submodule foreach git pull origin main && \
	$(shell go env GOPATH)/bin/hugo server -s hugo --disableFastRender

# Formats the code
pretty:
	gofmt -s -w .

# Saves changes to a file
diff:
	git diff > rego.diff

# Runs the application
run: build
	./${BINARY}

# Cleans the binary
clean:
	go clean
	rm ${BINARY}

# Flush the cache and any tests
flush:
	@/bin/bash -c 'if compgen -G "$$TMPDIR/rego_*" > /dev/null; then rm -rf $$TMPDIR/rego_*; fi'

# Replace every matching line in $(SRC_DIR) with NEW_COPYRIGHT
SRC_DIR := pkg
NEW_COPYRIGHT  := :Copyright: (c) 2025 by Gemini Software Services, LLC., see AUTHORS for more info

.PHONY: update-copyright
update-copyright:
	@echo "Updating copyright lines to $(NEW_COPYRIGHT)…"
	@find $(SRC_DIR) -type f -print0 | \
	  xargs -0 perl -pi -e 's/^:Copyright: \(c\) 202[0-9].*/'"$(NEW_COPYRIGHT)"'/mg'
	@echo "Done."
