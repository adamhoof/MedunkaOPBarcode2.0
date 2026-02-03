package utils

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// CreateSecureMQTTClient returns a fully configured client or panics.
// It does NOT connect (Connect is a separate step), but it prepares the configuration.
func CreateSecureMQTTClient(clientID string) mqtt.Client {
	opts := mqtt.NewClientOptions()
	protocol := GetEnvOrPanic("MQTT_PROTOCOL")
	host := GetEnvOrPanic("MQTT_HOST")
	port := GetEnvOrPanic("MQTT_PORT")
	opts.AddBroker(fmt.Sprintf("%s://%s:%s", protocol, host, port))
	opts.SetClientID(clientID)
	opts.SetUsername(ReadSecretOrFail("MQTT_USER_FILE"))
	opts.SetPassword(ReadSecretOrFail("MQTT_PASSWORD_FILE"))

	tlsConfig, err := generateTLSConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to generate TLS config: %v", err))
	}
	opts.SetTLSConfig(tlsConfig)

	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetCleanSession(false)
	opts.SetOrderMatters(false)

	return mqtt.NewClient(opts)
}

func generateTLSConfig() (*tls.Config, error) {
	clientCert, err := tls.LoadX509KeyPair(GetEnvOrPanic("TLS_CERT_PATH"), GetEnvOrPanic("TLS_KEY_PATH"))
	if err != nil {
		return nil, fmt.Errorf("failed to load client keypair: %w", err)
	}

	caBytes, err := os.ReadFile(GetEnvOrPanic("TLS_CA_PATH"))
	if err != nil {
		return nil, fmt.Errorf("failed to read CA cert: %w", err)
	}

	rootCAs := x509.NewCertPool()
	if ok := rootCAs.AppendCertsFromPEM(caBytes); !ok {
		return nil, fmt.Errorf("failed to parse CA cert PEM")
	}

	return &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      rootCAs,
	}, nil
}

// ConnectOrFail attempts to connect once and panics on failure.
func ConnectOrFail(client mqtt.Client) {
	token := client.Connect()
	if !token.WaitTimeout(5*time.Second) || token.Error() != nil {
		panic(fmt.Sprintf("CRITICAL: MQTT connection failed: %v", token.Error()))
	}
	fmt.Println("MQTT Connected")
}
