package utils

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"
	"time"
)

// CreateSecureHTTPClient returns an HTTP client configured with mTLS.
func CreateSecureHTTPClient(caPath, certPath, keyPath string) *http.Client {
	caCert, err := os.ReadFile(caPath)
	if err != nil {
		panic(err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		panic("failed to append CA certificate")
	}

	clientCert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		panic(err)
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:      certPool,
			Certificates: []tls.Certificate{clientCert},
		},
	}

	return &http.Client{
		Transport: transport,
		Timeout:   60 * time.Second,
	}
}
