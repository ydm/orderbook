OrderBook
========

OrderBook supports following operations:
1. Order submission
2. Cancellation of unexecuted orders
3. Querying
4. Order matching (on `AddOrder()`)
5. Order book snapshot

Files
------

```
orderbook
├── *.go    -- Order book implementation
├── cmd     -- (Example) REST API server
└── scripts -- REST API client scripts
```

#### TODO

Add more test cases and functionality:
- Cancel a partially executed order
- Create database abstraction
- Use a real database or at least SQLite3
