package caches

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/dgraph-io/ristretto"
	"github.com/vecosy/vecosy/v2/internal/utils"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
)

// ErrNotFound no public key found on the cache
var ErrNotFound = errors.New("no pubkey found")

type keyCache interface {
	StorePubKey(app *configrepo.ApplicationVersion, key *rsa.PublicKey) (*rsa.PublicKey, error)
	GetPubKey(app *configrepo.ApplicationVersion) (*rsa.PublicKey, error)
	GetOrSetPubKey(repo configrepo.Repo, app *configrepo.ApplicationVersion) (*rsa.PublicKey, error)
}

type keyCacheImpl struct {
	cache *ristretto.Cache
}

func (kc *keyCacheImpl) StorePubKey(app *configrepo.ApplicationVersion, key *rsa.PublicKey) (*rsa.PublicKey, error) {
	if !kc.cache.Set(kc.getKey(app), key, 1) {
		return key, errors.New("cannot store on the cache")
	}
	return key, nil
}

func (kc *keyCacheImpl) getKey(app *configrepo.ApplicationVersion) string {
	if app == nil {
		return ""
	}
	return fmt.Sprintf("%s-%s", app.AppName, app.AppVersion)
}

func (kc *keyCacheImpl) GetPubKey(app *configrepo.ApplicationVersion) (*rsa.PublicKey, error) {
	cacheVal, found := kc.cache.Get(kc.getKey(app))
	if !found {
		return nil, ErrNotFound
	}
	if pubkey, ok := cacheVal.(rsa.PublicKey); ok {
		return &pubkey, nil
	}
	return nil, ErrNotFound
}

func (kc *keyCacheImpl) GetOrSetPubKey(repo configrepo.Repo, app *configrepo.ApplicationVersion) (*rsa.PublicKey, error) {
	pubKey, err := kc.GetPubKey(app)
	if errors.Is(err, ErrNotFound) {
		pubKeyFile, err := repo.GetFile(app, "pub.key")
		if err != nil {
			return nil, err
		}
		pubKey, err = utils.BytesToPublicKey(pubKeyFile.Content)
		if err != nil {
			return nil, err
		}
		return kc.StorePubKey(app, pubKey)
	}
	return pubKey, nil
}
