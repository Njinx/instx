package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"time"

	"gitlab.com/Njinx/searx-space-autoselector/proxy/cert/util"
)

const CERT_FILE = "./certs/cert.pem"
const KEY_FILE = "./certs/key.pem"

const PROG_ID = "searx-space-autoselector"

type KeyPairBytes struct {
	Key  []byte
	Cert []byte
}

type KeyPair struct {
	Template x509.Certificate
	KeyPairBytes
}

func ValidateCerts() bool {
	certBytes, err := os.ReadFile(CERT_FILE)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false
		} else {
			log.Fatal("Could not read cert file: ", err)
		}
	}

	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return false
	}

	if cert.NotBefore.Unix() > time.Now().Unix() || cert.NotAfter.Unix() < time.Now().Unix() {
		return false
	}

	return true
}

func NewFromTemplate(template *x509.Certificate, keyPath string, certPath string) (KeyPairBytes, error) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	keyBytes := x509.MarshalPKCS1PrivateKey(key)
	certBytes, _ := x509.CreateCertificate(
		rand.Reader,
		template,
		template,
		&key.PublicKey,
		key)

	keyPairBytes := KeyPairBytes{
		Key:  keyBytes,
		Cert: certBytes,
	}

	certFile, err := os.OpenFile(CERT_FILE, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return keyPairBytes, err
	}
	defer certFile.Close()
	keyFile, err := os.OpenFile(KEY_FILE, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return keyPairBytes, err
	}
	defer keyFile.Close()

	err = pem.Encode(certFile, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	if err != nil {
		return keyPairBytes, err
	}
	err = pem.Encode(keyFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	})
	if err != nil {
		return keyPairBytes, err
	}

	return keyPairBytes, nil
}

func GenerateCerts() {
	hostname, err := os.Hostname()
	if err != nil {
		randomHostnameSize := new(big.Int).Lsh(big.NewInt(1), 16)
		hostnameRaw, err := rand.Int(rand.Reader, randomHostnameSize)
		if err != nil {
			hostname = "1"
		} else {
			hostname = base64.RawURLEncoding.EncodeToString(hostnameRaw.Bytes())
		}
	}

	organizationName := pkix.Name{
		Organization: []string{fmt.Sprintf("%s (%s)", PROG_ID, hostname)},
	}

	keyTemplate := x509.Certificate{
		SerialNumber:          util.GenerateSerialNumber(),
		Subject:               organizationName,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
		DNSNames:              []string{"localhost", "searx.local"},
	}

	_, err = NewFromTemplate(&keyTemplate, KEY_FILE, CERT_FILE)
	if err != nil {
		log.Fatal("Could not generate certificate: ", err)
	}
}
