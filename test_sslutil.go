// Copyright (c) 2022 MindStand Technologies, Inc
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package gogm

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

func createTempNeoKeypair() (string, error) {
	tempDir, err := ioutil.TempDir("gogm_test", "*")
	if err != nil {
		return "", err
	}

	ca := createCertificateAuthority()

	// CA Private Key

	caPk, err := createPrivateKey()
	if err != nil {
		return "", err
	}

	caPkPem, err := pemEncode(x509.MarshalPKCS1PrivateKey(caPk), "RSA PRIVATE KEY")
	if err != nil {
		return "", err
	}

	err = ioutil.WriteFile(filepath.Join(tempDir, "ca-private.key"), caPkPem, 0644)
	if err != nil {
		return "", err
	}

	// CA Cert

	caCertBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPk.PublicKey, caPk)
	if err != nil {
		return "", err
	}

	caCertPem, err := pemEncode(caCertBytes, "CERTIFICATE")
	if err != nil {
		return "", err
	}

	err = ioutil.WriteFile(filepath.Join(tempDir, "ca-public.crt"), caCertPem, 0644)
	if err != nil {
		return "", err
	}

	// Server Cert

	serverCert := createServerCertificate()

	// Private key

	pk, err := createPrivateKey()
	if err != nil {
		return "", err
	}

	pkPem, err := pemEncode(x509.MarshalPKCS1PrivateKey(pk), "RSA PRIVATE KEY")
	if err != nil {
		return "", err
	}

	err = ioutil.WriteFile(filepath.Join(tempDir, "private.key"), pkPem, 0644)
	if err != nil {
		return "", err
	}

	// Cert

	certBytes, err := x509.CreateCertificate(rand.Reader, serverCert, ca, &pk.PublicKey, caPk)
	if err != nil {
		return "", err
	}

	certPem, err := pemEncode(certBytes, "CERTIFICATE")
	if err != nil {
		return "", err
	}

	err = ioutil.WriteFile(filepath.Join(tempDir, "public.crt"), certPem, 0644)
	if err != nil {
		return "", err
	}

	return tempDir, nil
}

func cleanupTempNeoKeypair(tempDir string) error {
	return os.RemoveAll(tempDir)
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

func createPrivateKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 4096)
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

func pemEncode(payload []byte, payloadType string) ([]byte, error) {
	result := new(bytes.Buffer)
	err := pem.Encode(result, &pem.Block{
		Type:  payloadType,
		Bytes: payload,
	})
	if err != nil {
		return nil, err
	}
	return result.Bytes(), nil
}
