#!/bin/bash

go run server.go &
PARENT_PID=$!
sleep 1

# Find the actual Go binary started by `go run`
CHILD_PID=$(pgrep -P $PARENT_PID)

go test -count=1 .
TEST_EXIT_STATUS=$?

# Kill the compiled binary
kill $CHILD_PID
wait $CHILD_PID 2>/dev/null

exit $TEST_EXIT_STATUS
