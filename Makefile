test:
	go test
	/usr/bin/env bash e2e_test.sh

build:
	go build .
