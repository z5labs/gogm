package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strings"
	"time"
)

var basePath = os.Getenv("ROOT")

func main() {
	if !strings.HasSuffix(basePath, "/") {
		basePath += "/"
	}

	ca := createCertificateAuthority()
	caPrivateKey := createPrivateKey()

	caCertificateBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivateKey.PublicKey, caPrivateKey)
	panicOnError(err)
	panicOnError(ioutil.WriteFile(
		basePath+"ca-private.key",
		pemEncode(x509.MarshalPKCS1PrivateKey(caPrivateKey), "RSA PRIVATE KEY"),
		0644))
	fmt.Printf("wrote %sca-private.key\n", basePath)
	panicOnError(ioutil.WriteFile(
		basePath+"ca-public.crt",
		pemEncode(caCertificateBytes, "CERTIFICATE"),
		0644))
	fmt.Printf("wrote %sca-public.crt\n", basePath)

	serverCertificate := createServerCertificate()
	privateKey := createPrivateKey()
	certificateBytes, err := x509.CreateCertificate(rand.Reader, serverCertificate, ca, &privateKey.PublicKey, caPrivateKey)
	panicOnError(err)
	panicOnError(ioutil.WriteFile(
		basePath+"private.key",
		pemEncode(x509.MarshalPKCS1PrivateKey(privateKey), "RSA PRIVATE KEY"),
		0644))
	fmt.Printf("wrote %sprivate.key\n", basePath)
	panicOnError(ioutil.WriteFile(
		basePath+"public.crt",
		pemEncode(certificateBytes, "CERTIFICATE"),
		0644))
	fmt.Printf("wrote %spublic.crt\n", basePath)
}

func createCertificateAuthority() *x509.Certificate {
	return &x509.Certificate{
		SerialNumber:          big.NewInt(20211202),
		Subject:               caSubject(),
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
}

func createPrivateKey() *rsa.PrivateKey {
	result, err := rsa.GenerateKey(rand.Reader, 4096)
	panicOnError(err)
	return result
}

func createServerCertificate() *x509.Certificate {
	return &x509.Certificate{
		SerialNumber: big.NewInt(12022021),
		Subject:      caSubject(),
		DNSNames:     []string{"localhost"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
}

func caSubject() pkix.Name {
	return pkix.Name{
		Organization:  []string{"GoGM Test ORG"},
		Country:       []string{"US"},
		Province:      []string{"MD"},
		Locality:      []string{"Baltimore"},
		StreetAddress: []string{"n/a"},
		PostalCode:    []string{"n/a"},
	}
}

func pemEncode(payload []byte, payloadType string) []byte {
	result := new(bytes.Buffer)
	panicOnError(pem.Encode(result, &pem.Block{
		Type:  payloadType,
		Bytes: payload,
	}))
	return result.Bytes()
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
