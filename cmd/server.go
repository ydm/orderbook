package main

import (
	"fmt"

	"orderbook"
)

func main() {
	b := orderbook.NewBook()
	fmt.Printf("%v\n", b)
}
