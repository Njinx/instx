package ca

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"log"
	"time"

	"gitlab.com/Njinx/searx-space-autoselector/proxy/cert"
	"gitlab.com/Njinx/searx-space-autoselector/proxy/cert/util"
)

const CA_KEY_FILE = "./certs/ca/key.pem"
const CA_CERT_FILE = "./certs/ca/cert.pem"

type CertificateAuthority struct {
	cert.KeyPair
}

func (ca *CertificateAuthority) Sign(req *x509.CertificateRequest) {

}

func New() CertificateAuthority {
	var ca CertificateAuthority

	ca.Template = x509.Certificate{
		SerialNumber: util.GenerateSerialNumber(),
		Subject: pkix.Name{
			Organization:  []string{"searx-space-autoselector"},
			Country:       []string{"The Milky Way"},
			Province:      []string{"The Solar System"},
			Locality:      []string{"The Moon"},
			StreetAddress: []string{"Rupes Cauchy"},
			PostalCode:    []string{"Mare Tranquillitatis"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(20, 0, 0),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageServerAuth,
		},
	}

	keyPairBytes, err := cert.NewFromTemplate(&ca.Template, CA_KEY_FILE, CA_CERT_FILE)
	if err != nil {
		log.Fatal("Could not generate certificate: ", err)
	}
	ca.KeyPairBytes = keyPairBytes

	return ca
}
