package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	godiacritics "github.com/Regis24GmbH/go-diacritics"
	"github.com/adamhoof/MedunkaOPBarcode2.0/internal/database"
	"github.com/adamhoof/MedunkaOPBarcode2.0/internal/domain"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	maxPublishRetries = 3
)

// processJobs pulls from the channel forever.
func processJobs(db database.Handler, client mqtt.Client, jobs <-chan mqtt.Message, dbTimeout time.Duration) {
	for msg := range jobs {
		safeHandleRequest(db, client, msg, dbTimeout)
	}
}

// safeHandleRequest prevents a single bad message from crashing the app
func safeHandleRequest(db database.Handler, client mqtt.Client, msg mqtt.Message, dbTimeout time.Duration) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("CRITICAL: Worker panicked! Recovered: %v", r)
		}
	}()

	handleRequest(db, client, msg, dbTimeout)
}

// handleRequest is the actual business logic
func handleRequest(db database.Handler, client mqtt.Client, msg mqtt.Message, dbTimeout time.Duration) {
	var req domain.ProductDataRequest
	if err := json.Unmarshal(msg.Payload(), &req); err != nil {
		log.Printf("error unpacking payload: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	product, err := db.Fetch(ctx, req.Barcode)
	if err != nil {
		log.Printf("failed to fetch product %s: %v", req.Barcode, err)
		return
	}

	if !req.IncludeDiacritics {
		product.Name = godiacritics.Normalize(product.Name)
		product.Price = godiacritics.Normalize(product.Price)
	}

	respJSON, err := json.Marshal(product)
	if err != nil {
		log.Printf("failed to serialize response: %v", err)
		return
	}

	for i := 0; i < maxPublishRetries; i++ {
		token := client.Publish(req.ClientTopic, 1, false, respJSON)
		if token.WaitTimeout(1*time.Second) && token.Error() == nil {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}

	log.Printf("CRITICAL: Failed to publish response for %s", req.Barcode)
}
