#!/bin/bash

for i in {1..10}; do
  echo "Request $i session rate limit"
  curl -s -o /dev/null -w "HTTP %{http_code}\n" \
  -X POST \
  http://localhost:3000/api/v1/sessions
done