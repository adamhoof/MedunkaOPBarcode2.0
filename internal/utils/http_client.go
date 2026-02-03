package utils

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"
	"time"
)

// CreateSecureHTTPClient returns an HTTP client configured with the provided CA.
func CreateSecureHTTPClient(caPath string) *http.Client {
	caCert, err := os.ReadFile(caPath)
	if err != nil {
		panic(err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		panic("failed to append CA certificate")
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: certPool,
		},
	}

	return &http.Client{
		Transport: transport,
		Timeout:   60 * time.Second,
	}
}
