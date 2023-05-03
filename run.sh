#!/bin/sh
find . -type f -name '*.go' -o -name '*.sql' \
	| entr -rc go run main.go
