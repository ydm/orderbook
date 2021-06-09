all: orderbook server

orderbook:
	go build

server: clean
	go build cmd/server.go

clean:
	rm -f server

test:
	go test .
