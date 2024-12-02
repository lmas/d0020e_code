
# Run tests and log the test coverage
test:
	go test -v -coverprofile=".cover.out" "$$(go list ./... | grep -v /tmp)"

# Run benchmarking
bench:
	go test -cover -test.benchmem -bench=.

# Generate pretty coverage report
cover:
	go tool cover -html=".cover.out" -o="cover.html"

# Runs source code linters and catches common errors
lint:
	test -z $$(gofmt -l .) || (echo "Code isn't gofmt'ed!" && exit 1)
	go vet "$$(go list ./... | grep -v /tmp)"
	gosec -quiet -fmt=golint -exclude-dir="tmp" ./...
	# TODO: add cyclomatic complexxity tool?

# Updates 3rd party packages and tools
deps:
	go get -u "$$(go list ./... | grep -v /tmp)"
	go mod tidy
	go install github.com/securego/gosec/v2/cmd/gosec@latest

# Show documentation of public parts of package, in the current dir
docs:
	go doc -all
	# TODO: add class diagram tool

# Builds the binary, with debug symbol table and DWARF gen disabled for smaller bin
build:
	go build -ldflags "-s -w"
	# TODO: add raspberrypi build target

# Clean up built binary and other temporary files (ignores errors from rm)
clean:
	go clean
	rm .cover.out cover.html
	# TODO: add raspberrypi bins
