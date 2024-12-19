// You can edit this code!
// Click here and start typing.
package ca

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"
)

// This is a recursive tree of certificates, issued in-memory
type Certificate struct {
	// These are the attributes that get signed, kind of like the json body of a JWT
	// The common name determines file name. ie: CertAuthority.pem
	Template *x509.Certificate
	// This is the actual certificate bytes once signed, in openssl format
	BytesPem []byte
	// This is the parsed, in-memory key
	PrivateKey *rsa.PrivateKey
	// This is the key serialized out for openssl
	PrivateKeyPem []byte
	// The pem for the CSR that originated this cert
	CSRPem []byte
	// This is the list of everything this CA issued (if we are a CA!)
	Issued []*Certificate
	// This lets us do verification to the root
	Parent *Certificate
	// If we issued certs, then this is the last serial number used
	SerialNumber int64
}

func (subject *Certificate) DiskName(theDir string, thePem string) string {
	return fmt.Sprintf("%s/%s_%s.pem", theDir, subject.Template.Subject.CommonName, thePem)
}

func (subject *Certificate) Name() string {
	return subject.Template.Subject.CommonName
}

func (subject *Certificate) WriteToDisk(theDir string) error {
	var err error
	if len(subject.BytesPem) > 0 {
		err = os.WriteFile(
			subject.DiskName(theDir, "cert"),
			subject.BytesPem,
			0700,
		)
		if err != nil {
			return err
		}
	}
	if len(subject.CSRPem) > 0 {
		err = os.WriteFile(
			subject.DiskName(theDir, "csr"),
			subject.BytesPem,
			0700,
		)
		if err != nil {
			return err
		}
	}
	if len(subject.PrivateKeyPem) > 0 {
		err = os.WriteFile(
			subject.DiskName(theDir, "key"),
			subject.PrivateKeyPem,
			0700,
		)
		if err != nil {
			return err
		}
	}
	if len(subject.Issued) > 0 {
		for i := 0; i < len(subject.Issued); i++ {
			subject.Issued[i].WriteToDisk(theDir)
		}
	}
	return nil
}

func DecodeCSR(bytesPem []byte) (*x509.CertificateRequest, error) {
	block, _ := pem.Decode(bytesPem)
	return x509.ParseCertificateRequest(block.Bytes)
}

func RequestCertificate(subjectPrivateKey *rsa.PrivateKey, subjectCSRTemplate x509.CertificateRequest) ([]byte, error) {
	bytesASN1, err := x509.CreateCertificateRequest(
		rand.Reader,
		&subjectCSRTemplate,
		subjectPrivateKey,
	)
	if err != nil {
		return nil, err
	}
	bytesPem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: bytesASN1})
	return bytesPem, nil
}

func (subject *Certificate) SetPrivateKey(subjectPrivateKey *rsa.PrivateKey) {
	subject.PrivateKey = subjectPrivateKey
	bytesDer := x509.MarshalPKCS1PrivateKey(subject.PrivateKey)
	subject.PrivateKeyPem = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: bytesDer})
}

func (issuer *Certificate) IssueServer(
	csr *x509.CertificateRequest,
) (*Certificate, error) {
	// Make a keypair, but do NOT give it out to issuer.
	// Certificate signing requests go out, but private keys stay on our machine
	subjectPrivateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	subject, err := issuer.IssueServerForPublicKey(
		&subjectPrivateKey.PublicKey,
		csr,
	)
	if err != nil {
		return nil, err
	}

	// Our subject comes back as a *Certificate. We need to plug our private info into it now.
	subject.SetPrivateKey(subjectPrivateKey)
	return subject, err
}

func (issuer *Certificate) IssueRequest(subjectPublicKey *rsa.PublicKey, bytesCSRPem []byte) (*Certificate, error) {
	csr, err := DecodeCSR(bytesCSRPem)
	if err != nil {
		return nil, fmt.Errorf("unable to decode the CSR (certificate signing request) for: %v", err)
	}
	return issuer.IssueServerForPublicKey(subjectPublicKey, csr)
}

// the pkix.Name.CommonName is the DNS dial name that is expected
// but we also allow for localhost tunneling. the code would need to
// be expanded to use tunnels that are not on localhost
func (issuer *Certificate) IssueServerForPublicKey(
	subjectPublicKey *rsa.PublicKey,
	csr *x509.CertificateRequest,
) (*Certificate, error) {
	issuer.SerialNumber++
	subject := &Certificate{}
	subject.Template = &x509.Certificate{
		SerialNumber: big.NewInt(issuer.SerialNumber), // SerialNumber,Subject are a primary key
		Subject:      csr.Subject,
		NotBefore:    time.Now(),                   // clocks skew
		NotAfter:     time.Now().AddDate(10, 0, 0), // 10 year certificates
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageServerAuth,
		},
		IPAddresses: csr.IPAddresses, // allow tunnels via SAN
		DNSNames:    csr.DNSNames,
	}

	bytesDer, err := x509.CreateCertificate(
		rand.Reader,
		subject.Template,
		issuer.Template,
		subjectPublicKey, // our public key is signed into the certificate by the issuer
		issuer.PrivateKey,
	)
	if err != nil {
		return nil, err
	}
	subject.BytesPem = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: bytesDer})
	issuer.Issued = append(issuer.Issued, subject)
	return subject, nil
}

// Start here to generate a CA from which to run
//
//	IssueServer to make a Certificate Signing Request
//
// or RequestServer, from which to get a CA to give you the certificate
func NewCA(name pkix.Name) (*Certificate, error) {
	var err error

	subject := &Certificate{}
	issuer := subject

	// The private key derives the public key, so note that
	// the public key is a created part of the private key.
	issuer.PrivateKey, err = rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	// This issuer bumps up the serial number for every time its key is used.
	// So, this assumes that it is used SERIALLY to sign requests
	issuer.SerialNumber++
	subject.Template = &x509.Certificate{
		SerialNumber:          big.NewInt(issuer.SerialNumber), // SerialNumber,Subject are a primary key
		Subject:               name,
		NotBefore:             time.Now(),                   // clocks skew
		NotAfter:              time.Now().AddDate(10, 0, 0), // 10 year certificate
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		IsCA:                  true,
		BasicConstraintsValid: true,
	}
	// We now have a self-signed root cert for which to sign certificates
	// Note that the subject and issuer are a pointer to the same object,
	// which is what happens with self-signed CA
	bytesDer, err := x509.CreateCertificate(
		rand.Reader,
		subject.Template,              // This is info as the subject requested (also the issuer)
		issuer.Template,               // these pointers are the same!
		&subject.PrivateKey.PublicKey, // It is the subject's public key being signed ...
		issuer.PrivateKey,             // ... by the issuer's private key
	)
	if err != nil {
		return nil, err
	}
	// pem bytes are readable by openssl
	subject.BytesPem = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: bytesDer})
	bytesDer = x509.MarshalPKCS1PrivateKey(subject.PrivateKey)
	subject.PrivateKeyPem = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: bytesDer})

	// Get derBytes of the private key and store that too.
	return subject, nil
}
