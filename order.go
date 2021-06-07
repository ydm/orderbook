package orderbook

const (
	TypeLimit = iota
	TypeMarket
)

type Order struct {
	Type   int
	Symbol string
	ID     string
}
