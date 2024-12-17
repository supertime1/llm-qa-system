#!/bin/bash

# Generate Python code
python -m grpc_tools.protoc \
    -I../proto \
    --python_out=../llm-service/src \
    --grpc_python_out=../llm-service/src \
    ../proto/medical_qa.proto

# Generate Go code
protoc \
    -I../proto \
    --go_out=../backend-service/src \
    --go-grpc_out=../backend-service/src \
    ../proto/medical_service.proto