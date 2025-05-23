#!/bin/bash

go get -u ./...
go mod tidy
bash test.sh