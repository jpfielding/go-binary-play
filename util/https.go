package util

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"
)

// LoadHTTPClient creates the HTTPClient from files
func LoadHTTPClient(tlsVerify bool, cCertPath, cKeyPath, caCertPath string) (*http.Client, error) {
	var err error
	var caCert []byte
	if caCertPath != "" {
		if caCert, err = os.ReadFile(caCertPath); err != nil {
			return nil, err
		}
	}
	var certPEMBlock []byte
	var keyPEMBlock []byte
	if cCertPath != "" {
		if certPEMBlock, err = os.ReadFile(cCertPath); err != nil {
			return nil, err
		}
		if keyPEMBlock, err = os.ReadFile(cKeyPath); err != nil {
			return nil, err
		}
	}
	return NewHTTPClient(tlsVerify, certPEMBlock, keyPEMBlock, caCert)
}

// NewHTTPClient generates the sole connection per prrocess to outside services, TLS, mTLS w and w/o JWT
func NewHTTPClient(tlsVerify bool, cCert, cKey, caCert []byte) (*http.Client, error) {
	caCertPool := x509.NewCertPool()
	if caCert != nil {
		caCertPool.AppendCertsFromPEM(caCert)
	}
	certs := []tls.Certificate{}
	if cCert != nil {
		cert, err := tls.X509KeyPair(cCert, cKey)
		if err != nil {
			return nil, err
		}
		certs = append(certs, cert)
	}
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:            caCertPool,
				ClientCAs:          caCertPool,
				Certificates:       certs,
				InsecureSkipVerify: !tlsVerify,
			},
		},
	}, nil
}
