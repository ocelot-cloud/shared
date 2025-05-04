#!/bin/bash

set -e

PROJECT_DIR="$(pwd)"

cd "$PROJECT_DIR/main"
bash test.sh

cd "$PROJECT_DIR/utils"
go test .

cd "$PROJECT_DIR/validation"
go test .