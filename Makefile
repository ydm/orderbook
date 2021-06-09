server:
	go build cmd/server.go

orderbook:
	go build

all: orderbook test server

clean:
	rm -f server

test:
	go test ./...
