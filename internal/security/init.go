package security

import (
	"github.com/sirupsen/logrus"
	"github.com/vecosy/vecosy/v2/internal/caches"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
	"gopkg.in/square/go-jose.v2"
)

// CheckJwtToken check a jws token signature
func CheckJwtToken(repo configrepo.Repo, app *configrepo.ApplicationVersion, token string) error {
	log := logrus.WithField("method", "CheckJwtToken")
	repoPubKey, err := caches.KeyCache.GetOrSetPubKey(repo, app)
	if err != nil {
		log.Errorf("Error getting repo pub key:%s", err)
		return err
	}
	jws, err := jose.ParseSigned(token)
	if err != nil {
		log.Errorf("Error parsing jws:%s", err)
		return ErrAuthFailed
	}
	_, err = jws.Verify(repoPubKey)
	if err != nil {
		log.Errorf("Error verifying jws:%s", err)
		return ErrAuthFailed
	}
	return nil
}
