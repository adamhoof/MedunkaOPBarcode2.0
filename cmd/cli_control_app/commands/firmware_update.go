package commands

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func FirmwareUpdateServer(firmwarePath, certPath, keyPath, hostIP, port string) (string, func(), error) {
	if _, err := os.Stat(firmwarePath); os.IsNotExist(err) {
		return "", nil, fmt.Errorf("firmware file not found at path: %s", firmwarePath)
	}

	serverAddr := fmt.Sprintf("%s:%s", hostIP, port)
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
