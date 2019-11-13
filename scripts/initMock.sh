#!/usr/bin/env sh
mockgen -source ./pkg/configrepo/init.go -package mocks -destination ./mocks/configrepo_mock.go -imports version=github.com/hashicorp/go-version
mockgen -source ./internal/grpc/vecosy.pb.go -package grpc -destination ./internal/grpc/grpc_mock.go -self_package github.com/vecosy/vecosy/v2/internal/grpc
