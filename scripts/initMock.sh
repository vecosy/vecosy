#!/usr/bin/env sh
mockgen -source ./pkg/configrepo/init.go -package mocks -destination ./mocks/configrepo_mock.go -imports version=github.com/hashicorp/go-version
mockgen -source ./internal/grpcapi/vecosy.pb.go -package grpcapi -destination ./internal/grpcapi/grpcapi_mock.go -self_package github.com/vecosy/vecosy/v2/internal/grpcapi
mockgen -package mocks  -destination ./mocks/iris_mock.go github.com/kataras/iris Context
