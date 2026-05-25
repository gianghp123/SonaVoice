#!/bin/bash
for i in {1..70}; do
  echo "Request $i global rate limit"
  curl -s -o /dev/null -w "HTTP %{http_code}\n" http://localhost:3000/api/v1/health
done

