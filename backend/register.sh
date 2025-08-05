#!/bin/bash

# Set the API endpoint
URL="http://localhost:3000/register"

# Prepare the JSON payload
DATA='{
  "username": "somename",
  "password": "somePassword1",
  "email": "svenotimm@gmail.com"
}'

# Send the POST request and print the response
curl -X POST "$URL" \
  -H "Content-Type: application/json" \
  -d "$DATA"

