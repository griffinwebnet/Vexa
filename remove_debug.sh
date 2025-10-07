#!/bin/bash

# Remove all DEBUG printf statements from Go files
find api -name "*.go" -type f -exec sed -i '/fmt\.Printf.*DEBUG/d' {} \;

# Remove empty lines that might be left behind
find api -name "*.go" -type f -exec sed -i '/^[[:space:]]*$/N;/^\n$/d' {} \;

echo "Debug statements removed from all Go files"
