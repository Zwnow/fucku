#!/bin/bash

LOGIN_URL="http://localhost:3000/login"
LOGOUT_URL="http://localhost:3000/logout"
COOKIE_JAR="./cookies.txt"

# User login credentials
LOGIN_DATA='{
    "username": "somename",
    "email": "svenotimm@gmail.com",
    "password": "somePassword1"
}'

# Step 1: Log in and store cookies
curl -X POST "$LOGIN_URL" \
    -H "Content-Type: application/json" \
    -d "$LOGIN_DATA" \
    --cookie-jar "$COOKIE_JAR" \
    -i

# Extract session_token and csrf_token from cookie jar
SESSION_TOKEN=$(grep 'session_token' "$COOKIE_JAR" | awk '{print $7}')
CSRF_TOKEN=$(grep 'csrf_token' "$COOKIE_JAR" | awk '{print $7}')

if [[ -z "$SESSION_TOKEN" || -z "$CSRF_TOKEN" ]]; then
    echo "Failed to extract session_token or csrf_token from cookies."
    exit 1
fi

echo "Extracted session_token: $SESSION_TOKEN"
echo "Extracted csrf_token: $CSRF_TOKEN"

# Step 2: Call logout using the extracted tokens
curl -X POST "$LOGOUT_URL" \
    --cookie "$COOKIE_JAR" \
    -H "X-CSRF-Token: $CSRF_TOKEN" \
    -H "Content-Type: application/json" \
    -i

