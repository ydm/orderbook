.PHONY: help
help:
	@echo 'Management commands:'
	@echo
	@echo 'Usage:'
	@echo '    make help            Print this help message.'
	@echo '    make all             Lint and build.'
	@echo '    make clean           Clean directory tree.'
	@echo '    make fix             Fix small linting problems.'
	@echo '    make lint            Run static analysis on source code.'
	@echo '    make build           Compile project.'
	@echo '    make test            Run tests.'
	@echo

.PHONY: all
all: clean lint build

.PHONY: clean
clean:
	rm -f server

.PHONY: fix
fix:
	golangci-lint run --fix

.PHONY: lint
lint:
	golangci-lint run

.PHONY: build
build:
	go build
	go build cmd/server.go

.PHONY: test
test:
	go test .
