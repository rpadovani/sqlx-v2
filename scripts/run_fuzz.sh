#!/bin/bash

# Copyright 2026 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

# NOTE: -msan requires clang
export CC=clang
export CXX=clang++
export GOMEMLIMIT=800MiB
export GOMAXPROCS=2


DURATION=${1:-5m}

echo "==========================================="
echo "Running Target 1 Fuzzer: reflectx.TypeMap"
echo "==========================================="

echo "--- Running with -race ---"
go test -v -fuzz=FuzzTypeMap -fuzztime=$DURATION -race ./internal/reflectx/...

echo ""
echo "--- Running with -msan ---"
go test -v -fuzz=FuzzTypeMap -fuzztime=$DURATION -msan ./internal/reflectx/...

echo ""
echo "==========================================="
echo "Running Target 2 Fuzzer: Row Scan Bounds"
echo "==========================================="

echo "--- Running with -race ---"
go test -v -fuzz=FuzzRowScanBounds -fuzztime=$DURATION -race .

echo ""
echo "--- Running with -msan ---"
go test -v -fuzz=FuzzRowScanBounds -fuzztime=$DURATION -msan .

echo ""
echo "Fuzzing complete."
