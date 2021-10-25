help:
	@echo 'Management commands:'
	@echo
	@echo 'Usage:'
	@echo '    make help            Print this.'
	@echo '    make all             Make lint -> orderbook -> server.'
	@echo '    make clean           Clean directory tree.'
	@echo '    make fix             Fix small linting problems.'
	@echo '    make lint            Run static analysis on source code.'
	@echo '    make orderbook       Compile project.'
	@echo '    make server          Compile server executable.'
	@echo '    make test            Run tests.'
	@echo

all: lint orderbook server

clean:
	rm -f server

fix:
	golangci-lint run --fix

lint:
	golangci-lint run

orderbook:
	go build

server: clean
	go build cmd/server.go

test:
	go test .

.PHONY: all clean help lint orderbok server test
