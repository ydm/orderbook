OrderBook
========

It supports:
1. Order submitting
2. Cancellation
3. Querying
4. Order matching (on AddOrder)
5. Order book snapshot

Files
------

```
orderbook
├── *.go    -- Order book implementation
├── cmd     -- (Example) REST API server
└── scripts -- REST API client scripts
```

### TODO:

1. What will happen if there are the following three bid levels:
   - price=100, quantity=3
   -  price=99, quantity=2
   -  price=98, quantity=5

and a sell limit order for quantity of 3+ is placed at 99?  It has to
sell at 100 first, not at 99. Also, what's left, should be matched
against the next acceptable level then.  I don't remember if I wrote
it this way.  If I didn't, it's a bug.
