#!/usr/bin/bash

protoc  --go_out=. --go-grpc_out=. pkg/api/proto/*.proto
