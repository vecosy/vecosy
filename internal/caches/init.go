package caches

import (
	"github.com/dgraph-io/ristretto"
	"github.com/sirupsen/logrus"
)

var KeyCache keyCache

func init() {
	var err error
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 10000,
		MaxCost:     2e+8,
		BufferItems: 64,
	})
	KeyCache = &keyCacheImpl{cache: cache}
	if err != nil {
		logrus.Fatalf("Error initializing keyCache:%s", err)
	}
}
