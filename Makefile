all: lint orderbook server

clean:
	rm -f server

help:
	@echo 'Management commands:'
	@echo
	@echo 'Usage:'
	@echo '    make help            Print this.'
	@echo '    make clean           Clean directory tree.'
	@echo '    make lint            Run static analysis on source code.'
	@echo '    make orderbook       Compile project.'
	@echo '    make server          Compile server executable.'
	@echo '    make test            Run tests.'
	@echo

lint:
	golangci-lint run

orderbook:
	go build

server: clean
	go build cmd/server.go

test:
	go test .

.PHONY: all clean help lint orderbok server test
