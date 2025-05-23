#!/bin/bash

set -e

PROJECT_DIR="$(pwd)"

cd "$PROJECT_DIR/main"
go run .

cd "$PROJECT_DIR/utils"
go test .

cd "$PROJECT_DIR/validation"
go test .