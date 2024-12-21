#!/bin/bash

# Generate Python code for LLM service
python -m grpc_tools.protoc \
    -I./proto \
    --python_out=./llm-service/src \
    --grpc_python_out=./llm-service/src \
    ./proto/medical_service.proto

# Generate Go code for backend service
protoc \
    -I./proto \
    --go_out=./backend-service/src/proto --go_opt=paths=source_relative \
    --go-grpc_out=./backend-service/src/proto --go-grpc_opt=paths=source_relative \
    ./proto/medical_service.proto

# Ensure proper module path in generated files
find ./backend-service/src/proto -type f -name "*.go" -exec \
    sed -i.bak 's|package proto|package proto|' {} \;

echo "Proto generation complete!"