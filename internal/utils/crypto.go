package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gopkg.in/square/go-jose.v2"
	"log"
	"testing"
)

func BytesToPublicKey(pub []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pub)
	enc := x509.IsEncryptedPEMBlock(block)
	b := block.Bytes
	var err error
	if enc {
		log.Println("is encrypted pem block")
		b, err = x509.DecryptPEMBlock(block, nil)
		if err != nil {
			return nil, err
		}
	}
	ifc, err := x509.ParsePKIXPublicKey(b)
	if err != nil {
		return nil, err
	}
	key, ok := ifc.(*rsa.PublicKey)
	if !ok {
		return nil, err
	}
	return key, nil
}

func PublicKeyToBytes(pub *rsa.PublicKey) []byte {
	pubASN1, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		logrus.Error(err)
	}

	pubBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubASN1,
	})

	return pubBytes
}

func GenerateKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	return privKey, &privKey.PublicKey, nil
}

func GenJwsFromPrivateKey(t *testing.T, privKey *rsa.PrivateKey, appName string) *jose.JSONWebSignature {
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.PS512, Key: privKey}, nil)
	assert.NoError(t, err)
	jws, err := signer.Sign([]byte(appName))
	assert.NoError(t, err)
	return jws
}
