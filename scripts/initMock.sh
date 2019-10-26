#!/usr/bin/env sh
mockgen -source ./pkg/configrepo/init.go -package mocks -destination ./mocks/configrepo_mock.go -imports version=github.com/hashicorp/go-version