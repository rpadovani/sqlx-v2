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

# Require clang for msan
export CC=clang
export CXX=clang++

echo "==========================================="
echo "Running All Tests with Race Sanitizer"
echo "==========================================="
for mod in $(find . -name go.mod -type f); do
  dir=$(dirname "$mod")
  echo "Testing module in $dir"
  (cd "$dir" && go test -v -race ./...)
done

echo "==========================================="
echo "Running All Tests with Memory Sanitizer (MSAN)"
echo "==========================================="
for mod in $(find . -name go.mod -type f); do
  dir=$(dirname "$mod")
  echo "Testing module in $dir"
  (cd "$dir" && go test -v -msan ./...)
done

echo "Success: All tests passed with -race and -msan."
