all: orderbook test server

orderbook:
	go build

server:
	go build cmd/server.go

clean:
	rm -f server

test:
	go test ./...
