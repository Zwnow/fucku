#!/bin/bash

URL="http://localhost:3000/logout"
DATA='{
    "username": "somename",
    "email": "svenotimm@gmail.com",
    "password": "somePassword1"
}'

curl -X POST "$URL" \
    -H "Content-Type: application/json" \
    -d "$DATA" \
    -i
