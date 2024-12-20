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

# Ensure proper module path in medical_qa.proto
sed -i '' 's|// option go_package.*|option go_package = "github.com/supertime1/llm-qa-system/backend-service/src/proto";|' ./proto/medical_service.proto