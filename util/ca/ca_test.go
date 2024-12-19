package ca_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"net"
	"os"
	"testing"

	"github.com/jpfielding/go-binary-play/util/ca"
)

func TestIssueServerCert(t *testing.T) {
	// Generate a certificate authority
	certAuthorityName := "CertAuthority"
	t.Logf("Create a CA %s\n", certAuthorityName)
	caCertificate, err := ca.NewCA(
		pkix.Name{
			Country:      []string{"US"},
			Organization: []string{"Leidos"},
			Locality:     []string{"Morgantown"},
			CommonName:   certAuthorityName, // It is important to have a cn that is a dns dial name
		},
	)
	if err != nil {
		t.Logf("Unable to generate %s cert ca: %v", certAuthorityName, err)
		t.FailNow()
	}

	t.Run("lan1bags", func(t *testing.T) {
		// Generate a certificate directly, ignoring the CSR step
		lane1bagsName := "lane1bags.mgw-airport.com"
		t.Logf("Create %s server cert", lane1bagsName)
		lane1bags, err := caCertificate.IssueServer(
			&x509.CertificateRequest{
				Subject: pkix.Name{
					Country:      []string{"US"},
					Organization: []string{"Leidos"},
					Locality:     []string{"Morgantown"},
					CommonName:   lane1bagsName,
				},
				IPAddresses: []net.IP{net.ParseIP("127.0.0.1")}, // allow tunnels via SAN
				DNSNames:    []string{lane1bagsName, "localhost"},
			},
		)
		if err != nil {
			t.Logf("Unable to generate server %s cert: %v", lane1bagsName, err)
			t.FailNow()
		}
		t.Logf("issued cert for %s", lane1bags.Template.Subject.CommonName)
	})

	t.Run("lane1body", func(t *testing.T) {
		// Generate a keypair, then a CSR (certificate signing request) with it
		// The keypair stays LOCAL to our machine. Only the public key is given to the CA
		lane1bodyName := "lane1body.mgw-airport.com"
		t.Logf("Create a CSR for %s %s to issue", lane1bodyName, certAuthorityName)
		subjectPrivateKey, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			t.Logf("unable to generate a keypair for %s", lane1bodyName)
			t.FailNow()
		}
		// This serialized CSR (a pem file, is sent out to the CA for signing
		bytesCSRPem, err := ca.RequestCertificate(
			subjectPrivateKey, // we locally generate a CSR to send to the CA later
			x509.CertificateRequest{
				Subject: pkix.Name{
					Country:      []string{"US"},
					Organization: []string{"Leidos"},
					Locality:     []string{"Morgantown"},
					CommonName:   lane1bodyName,
				},
				IPAddresses: []net.IP{net.ParseIP("127.0.0.1")}, // allow tunnels via SAN
				DNSNames:    []string{lane1bodyName, "localhost"},
			},
		)

		// When the CA gets the CSR bytes, there is a decision on whether to issue it
		// It is written into subject when issued.
		subject, err := caCertificate.IssueRequest(
			&subjectPrivateKey.PublicKey,
			bytesCSRPem,
		)
		if err != nil {
			t.Logf("unable to issue the CSR (certificate signing request) for %s: %v", lane1bodyName, err)
			t.FailNow()
		}
		//The subject can pair it with his private key
		subject.SetPrivateKey(subjectPrivateKey)

		t.Logf("issued cert for %s", subject.Template.Subject.CommonName)
	})

	// These are being written out to ensure that it doesn't blow up.
	// You can check these with openssl to independently verify them:
	//   cat /tmp/prosight/CertAuthority_cert.pem | openssl x509 -text
	//   cat /tmp/prosight/CertAuthority_key.pem
	toDir := "/tmp/prosight"
	t.Logf("Persist ca and server to disk in %s", toDir)
	os.MkdirAll(toDir, 0700)
	caCertificate.WriteToDisk(toDir)
}
