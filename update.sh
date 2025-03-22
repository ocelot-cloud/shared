#!/bin/bash

go mod edit -go=$(go version | awk '{print $3}' | cut -c3-)
go get -u ./...
go mod tidy