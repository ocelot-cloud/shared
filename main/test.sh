#!/bin/bash

go run server.go &
sleep 1
go test -count=1 .
TEST_EXIT_STATUS=$?
pkill -f "go run server.go"
exit $TEST_EXIT_STATUS