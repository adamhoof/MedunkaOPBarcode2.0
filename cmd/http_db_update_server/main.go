package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/adamhoof/MedunkaOPBarcode2.0/internal/database"
	"github.com/adamhoof/MedunkaOPBarcode2.0/internal/utils"
)

func main() {
	postgreSQLHandler, err := database.New("postgres")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if closeErr := postgreSQLHandler.Close(); closeErr != nil {
			log.Printf("failed to close database connection: %s", closeErr)
		}
	}()

	jobStore := &sync.Map{}

	maxUploadSize := utils.GetEnvAsInt64("HTTP_MAX_UPLOAD_SIZE")

	http.HandleFunc(utils.GetEnvOrPanic("HTTP_SERVER_UPDATE_ENDPOINT"), HandleDatabaseUpdate(postgreSQLHandler, jobStore, maxUploadSize))
	http.HandleFunc(utils.GetEnvOrPanic("HTTP_SERVER_UPDATE_STATUS_ENDPOINT"), HandleJobStatus(jobStore))

	host := utils.GetEnvOrPanic("HTTP_SERVER_HOST")
	port := utils.GetEnvOrPanic("HTTP_SERVER_PORT")
	serverCertPath := utils.GetEnvOrPanic("TLS_SERVER_CERT_PATH")
	serverKeyPath := utils.GetEnvOrPanic("TLS_SERVER_KEY_PATH")
	caPath := utils.GetEnvOrPanic("TLS_CA_PATH")

	caCert, err := os.ReadFile(caPath)
	if err != nil {
		log.Fatalf("failed to load CA cert: %v", err)
	}

	clientCAPool := x509.NewCertPool()
	if ok := clientCAPool.AppendCertsFromPEM(caCert); !ok {
		log.Fatal("failed to parse CA cert")
	}

	tlsConfig := &tls.Config{
		ClientAuth: tls.RequireAndVerifyClientCert,
		ClientCAs:  clientCAPool,
		MinVersion: tls.VersionTLS12,
	}

	log.Printf("Starting server on %s:%s", host, port)

	server := &http.Server{
		Addr:      fmt.Sprintf("%s:%s", host, port),
		TLSConfig: tlsConfig,
	}

	err = server.ListenAndServeTLS(serverCertPath, serverKeyPath)
	if err != nil {
		log.Fatal("unable to start")
	}
}
