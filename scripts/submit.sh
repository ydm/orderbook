#!/bin/bash

# This script submits an order.

# Side:
#   0 - Buy
#   1 - Sell

# Type:
#   0 - Limit
#   1 - Market

SERVER=127.0.0.1:7701

read -d '' BODY << EOF
{
    "side": 1,
    "quantity": "10",
    "price": "1000",
    "id": "something",
    "type": 0
}
EOF

curl -X POST \
     -H "Content-Type: application/json" \
     -d "$BODY" \
     $SERVER/orders/
echo
