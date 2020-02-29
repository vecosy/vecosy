package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gopkg.in/square/go-jose.v2"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"testing"
	"time"
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

func GenerateCertificate() ([]byte, []byte, error) {
	privKey, pubKey, err := GenerateKeyPair()
	if err != nil {
		return nil, nil, err
	}
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, err
	}
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Vecosy test"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(1 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1)},
	}
	cerBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, pubKey, privKey)
	if err != nil {
		return nil, nil, err
	}
	pemPrivKey := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privKey),
		},
	)
	pemCert := pem.EncodeToMemory(
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cerBytes,
		},
	)

	return pemCert, pemPrivKey, nil
}

func GenerateCertificateFiles() (string, string, error) {
	cert, certKey, err := GenerateCertificate()
	certFile, err := ioutil.TempFile("", "cert")
	if err != nil {
		return "", "", err
	}
	_, err = certFile.Write(cert)
	if err != nil {
		return "", "", err
	}
	_ = certFile.Close()

	certKeyFile, err := ioutil.TempFile("", "certKey")
	if err != nil {
		return "", "", err
	}

	_, err = certKeyFile.Write(certKey)
	if err != nil {
		return "", "", err
	}
	_ = certKeyFile.Close()
	return certFile.Name(), certKeyFile.Name(), err
}
