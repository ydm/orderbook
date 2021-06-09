#!/bin/bash

# This script cancels an order.

if [ -z "$1" ] ; then
    echo "usage: $0 <id>"
    exit 1
fi

SERVER=127.0.0.1:7701

curl -X GET \
     -H "Content-Type: application/json" \
     -d "$BODY" \
     $SERVER/orders/$1
echo
