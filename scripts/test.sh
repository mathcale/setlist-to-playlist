#!/bin/sh

set -e
echo "mode: set" > coverage.tmp

for d in $(go list ./... | grep -v vendor); do
  go test -coverprofile=profile.out -covermode=set $d

  if [ -f profile.out ]; then
    tail -n +2 profile.out >>coverage.tmp
    rm profile.out
  fi
done

cat coverage.tmp | grep -v "/internal/pkg/mocks" > coverage.txt
rm coverage.tmp

go tool cover -html=coverage.txt
