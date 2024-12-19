package util_test

import (
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jpfielding/go-binary-play/util"
	"github.com/jpfielding/go-binary-play/util/ca"
	"github.com/stretchr/testify/assert"
)

// If we want to test client mTLS setup, we can use this to make mTLS server certs
func CreateMTLSContext(verify bool, cert, key, trust string) (*tls.Config, error) {
	caCert, err := os.ReadFile(trust)
	if err != nil {
		return nil, fmt.Errorf("unable to read tls cert config %s: %v", trust, err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	certParsed, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return nil, fmt.Errorf("unable to read (cert,key) pair (%s,%s): %v", cert, key, err)
	}

	tlsConfig := &tls.Config{
		RootCAs:            caCertPool,
		ClientCAs:          caCertPool,
		Certificates:       []tls.Certificate{certParsed},
		ClientAuth:         tls.RequireAndVerifyClientCert,
		InsecureSkipVerify: !verify, // inverting this arg is a little dangerous. use care.
	}
	return tlsConfig, nil
}

func generateCA(t *testing.T, nm string) *ca.Certificate {
	t.Logf("Create a CA %s\n", nm)
	cert, err := ca.NewCA(
		pkix.Name{
			Country:      []string{"US"},
			Organization: []string{"Leidos"},
			Locality:     []string{"Morgantown"},
			CommonName:   nm, // It is important to have a cn that is a dns dial name
		},
	)
	if err != nil {
		t.Logf("Unable to generate %s cert ca: %v", cert.Name(), err)
		t.FailNow()
	}
	return cert
}

// We allow certs to be used for both client and server, to simplify things
// We assume that these are MACHINE clients that still have DNS names.
func generateCert(
	t *testing.T,
	nm string,
	caCertificate *ca.Certificate,
) *ca.Certificate {
	// Generate a certificate directly, ignoring the CSR step
	t.Logf("Create %s server cert", nm)
	cert, err := caCertificate.IssueServer(
		&x509.CertificateRequest{
			Subject: pkix.Name{
				Country:      []string{"US"},
				Organization: []string{"Leidos"},
				Locality:     []string{"Morgantown"},
				CommonName:   nm,
			},
			IPAddresses: []net.IP{net.ParseIP("127.0.0.1")}, // allow tunnels via SAN
			DNSNames:    []string{nm, "localhost"},
		},
	)
	if err != nil {
		t.Logf("Unable to generate server %s cert: %v", nm, err)
		t.FailNow()
	}
	t.Logf("issued cert for %s", cert.Template.Subject.CommonName)
	return cert
}

func TestTLS(t *testing.T) {
	// run without timeout or in debug mode since cert gen takes time
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	// write certs temporarily into /tmp/prosight
	x509Path := "/tmp/prosight/ca" + "/" + uuid.New().String()
	err := os.MkdirAll(x509Path, 0700)
	assert.Nil(t, err)

	// issue a ca to sign the certs
	serverCA := generateCA(t, "serverCA")
	serverCert := generateCert(t, "server", serverCA)
	serverCA.WriteToDisk(x509Path)
	// issue a client ca to sign client certs
	clientCA := generateCA(t, "clientCA")
	clientCert := generateCert(t, "client", clientCA)
	clientCA.WriteToDisk(x509Path)
	//set the variables
	sCert := serverCert.DiskName(x509Path, "cert")
	sKey := serverCert.DiskName(x509Path, "key")
	sTrust := clientCA.DiskName(x509Path, "cert")
	cCert := clientCert.DiskName(x509Path, "cert")
	cKey := clientCert.DiskName(x509Path, "key")
	cTrust := serverCA.DiskName(x509Path, "cert")

	// create the server
	var tlsOK http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!"))
	}

	serverTLSConfig, err := CreateMTLSContext(
		true,
		sCert,
		sKey,
		sTrust,
	)
	if err != nil {
		t.Errorf("Failed to create test TLS context: %v", err)
		t.FailNow()
	}

	svr := &http.Server{
		Addr:      "localhost:0",
		Handler:   tlsOK,
		TLSConfig: serverTLSConfig,
	}

	ln, err := tls.Listen("tcp", svr.Addr, serverTLSConfig)
	assert.Nil(t, err)
	go func() {
		if err := svr.Serve(ln); err != nil && err != http.ErrServerClosed {
			assert.Nil(t, err)
		}
		defer ln.Close()
	}()

	// create the client
	cl, err := util.LoadHTTPClient(true, cCert, cKey, cTrust)
	assert.Nil(t, err)

	// request with host = cn=<server>
	port := ln.Addr().(*net.TCPAddr).Port
	u := fmt.Sprintf("https://localhost:%d", port)
	req, err := http.NewRequest("GET", u, nil)
	assert.Nil(t, err)

	// send the request
	res, err := cl.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}
