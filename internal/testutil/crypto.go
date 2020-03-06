package testutil

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
	"math/big"
	"net"
	"testing"
	"time"
)

// GenerateCertificate TEST-ONLY: generate a private key and a certificate
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

// GenerateCertificateFiles TEST ONLY: will generate a test certificate/priv key and store it on files
func GenerateCertificateFiles() (string, string, error) {
	cert, certKey, err := GenerateCertificate()
	if err != nil {
		return "", "", err
	}
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

// GenerateKeyPair TEST ONLY: generate a new RSA key pairs
func GenerateKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	return privKey, &privKey.PublicKey, nil
}

// GenJwsFromPrivateKey TEST ONLY: generate a new JWS token signed by a privKey
func GenJwsFromPrivateKey(t *testing.T, privKey *rsa.PrivateKey, appName string) *jose.JSONWebSignature {
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.PS512, Key: privKey}, nil)
	assert.NoError(t, err)
	jws, err := signer.Sign([]byte(appName))
	assert.NoError(t, err)
	return jws
}

// PublicKeyToBytes TEST ONLY:marshall an rsa.PublicKey to an array of bytes
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
