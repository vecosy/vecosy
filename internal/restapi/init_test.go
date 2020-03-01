package restapi

import (
	"crypto/rsa"
	"fmt"
	"github.com/gavv/httpexpect"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/vecosy/vecosy/v2/internal/utils"
	"github.com/vecosy/vecosy/v2/mocks"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
	"gopkg.in/square/go-jose.v2"
	"testing"
)

func applySecurity(t *testing.T, privKey *rsa.PrivateKey, req *httpexpect.Request, repo *mocks.MockRepo, app *configrepo.ApplicationVersion) {
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.PS512, Key: privKey}, nil)
	assert.NoError(t, err)
	jws, err := signer.Sign([]byte("TestApp"))
	assert.NoError(t, err)
	req.WithHeader("Authorization", fmt.Sprintf("Bearer %s", jws.FullSerialize()))
	repo.EXPECT().GetFile(app, "pub.key").Return(&configrepo.RepoFile{
		Version: uuid.New().String(),
		Content: utils.PublicKeyToBytes(&privKey.PublicKey),
	}, nil)
}
