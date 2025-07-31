#!/bin/bash

echo "Building xfsquota..."
go build -o xfsquota ./cmd/xfsquota

if [ $? -eq 0 ]; then
    echo "Build successful!"
    echo "Usage: ./xfsquota --help"
else
    echo "Build failed!"
    exit 1
fi 