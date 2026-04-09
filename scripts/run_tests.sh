#!/bin/bash
set -e

export CC=clang
export CXX=clang++

cur_go=$(go env GOVERSION | sed 's/^go//' | cut -d. -f1,2)

echo "==========================================="
echo "Running All Tests with Race Sanitizer"
echo "==========================================="
for mod in $(find . -name go.mod -type f); do
  dir=$(dirname "$mod")
  mod_go=$(grep -m1 -oE '^go [0-9]+\.[0-9]+' "$mod" | cut -d' ' -f2 || echo "1.0")
  if awk -v c="$cur_go" -v m="$mod_go" 'BEGIN{ if (c < m) exit 1 }'; then
    echo "Testing module in $dir"
    (cd "$dir" && go test -v -race ./...)
  else
    echo "Skipping module in $dir (requires Go $mod_go, running Go $cur_go)"
  fi
done

echo "==========================================="
echo "Running All Tests with Memory Sanitizer (MSAN)"
echo "==========================================="
for mod in $(find . -name go.mod -type f); do
  dir=$(dirname "$mod")
  mod_go=$(grep -m1 -oE '^go [0-9]+\.[0-9]+' "$mod" | cut -d' ' -f2 || echo "1.0")
  if awk -v c="$cur_go" -v m="$mod_go" 'BEGIN{ if (c < m) exit 1 }'; then
    echo "Testing module in $dir"
    (cd "$dir" && go test -v -msan ./...)
  else
    echo "Skipping module in $dir (requires Go $mod_go, running Go $cur_go)"
  fi
done

echo "Success: All tests passed with -race and -msan."
