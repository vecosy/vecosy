#!/usr/bin/env sh

protoc -I ./ --go_out=plugins=grpc:internal/grpc vecosy.proto