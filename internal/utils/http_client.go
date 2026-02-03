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

/*package utils

import (
"crypto/tls"
"crypto/x509"
"fmt"
"net/http"
"os"
"time"
)

// CreateSecureHTTPClient returns an HTTP client configured with the provided CA.
func CreateSecureHTTPClient(caPath, certPath, keyPath string) *http.Client {
	caCert, err := os.ReadFile(caPath)
	if err != nil {
		panic(fmt.Sprintf("Failed to read CA cert at %s: %v", caPath, err))
	}
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		panic("Failed to parse CA cert")
	}

	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		panic(fmt.Sprintf("Failed to load client certificate/key: %v", err))
	}

	tlsConfig := &tls.Config{
		RootCAs:      caCertPool,
		Certificates: []tls.Certificate{cert},
	}

	return &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}
}
*/
