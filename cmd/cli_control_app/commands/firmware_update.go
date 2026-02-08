package commands

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

func FirmwareUpdateServer(firmwarePath, certPath, keyPath string) (string, func(), error) {
	if _, err := os.Stat(firmwarePath); os.IsNotExist(err) {
		return "", nil, fmt.Errorf("firmware file not found at path: %s", firmwarePath)
	}

	ip, err := getLocalIP()
	if err != nil {
		return "", nil, fmt.Errorf("could not determine local IP: %v", err)
	}

	port := "9000"
	serverAddr := fmt.Sprintf("%s:%s", ip, port)
	downloadUrl := fmt.Sprintf("https://%s/fw", serverAddr)

	log.Printf("Starting temporary firmware server at %s", downloadUrl)

	mux := http.NewServeMux()
	mux.HandleFunc("/fw", func(w http.ResponseWriter, r *http.Request) {
		remoteAddr := r.RemoteAddr
		log.Printf(">> STATION CONNECTED: %s - Starting download...", remoteAddr)

		start := time.Now()
		fmt.Printf("serving firmware file: %s\n", firmwarePath)
		http.ServeFile(w, r, firmwarePath)

		log.Printf("<< STATION FINISHED: %s - Downloaded in %v", remoteAddr, time.Since(start))
	})

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	serverErr := make(chan error, 1)
	go func() {
		if err := server.ListenAndServeTLS(certPath, keyPath); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return "", nil, fmt.Errorf("server failed to start: %v", err)
	case <-time.After(200 * time.Millisecond):
	}

	waitAndStop := func() {
		fmt.Println("Server is OPEN. Waiting 10s for stations to connect...")
		time.Sleep(10 * time.Second)
		fmt.Println("Window CLOSED. Waiting for active downloads to finish...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Shutdown forced (timeout): %v", err)
		} else {
			log.Println("All stations finished. Server stopped gracefully.")
		}
	}

	return downloadUrl, waitAndStop, nil
}

func getLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}
