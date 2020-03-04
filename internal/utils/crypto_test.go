package utils

import (
	"github.com/stretchr/testify/assert"
	"github.com/vecosy/vecosy/v2/internal/testutil"
	"testing"
)

func TestPublicKeyConversions(t *testing.T) {
	check := assert.New(t)
	privKey, pubKey, err := testutil.GenerateKeyPair()
	check.NoError(err)
	check.NotNil(privKey)
	check.NotNil(pubKey)
	pubKeyBytes := testutil.PublicKeyToBytes(pubKey)
	check.NotEmpty(pubKeyBytes)
	readedPubKey, err := BytesToPublicKey(pubKeyBytes)
	check.NoError(err)
	check.EqualValues(readedPubKey, pubKey)

	readedPubKey, err = BytesToPublicKey([]byte("notValidPubKey"))
	check.Error(err)
	check.Nil(readedPubKey)
}

