package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	godiacritics "github.com/Regis24GmbH/go-diacritics"
	"github.com/adamhoof/MedunkaOPBarcode2.0/internal/database"
	"github.com/adamhoof/MedunkaOPBarcode2.0/internal/domain"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	maxPublishRetries = 3
)

func processJobs(db database.Handler, client mqtt.Client, jobs <-chan mqtt.Message, dbTimeout time.Duration) {
	for msg := range jobs {
		safeHandleRequest(db, client, msg, dbTimeout)
	}
}

func safeHandleRequest(db database.Handler, client mqtt.Client, msg mqtt.Message, dbTimeout time.Duration) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("CRITICAL: Worker panicked! Recovered: %v", r)
		}
	}()

	handleRequest(db, client, msg, dbTimeout)
}

func handleRequest(db database.Handler, client mqtt.Client, msg mqtt.Message, dbTimeout time.Duration) {
	parts := strings.Split(msg.Topic(), "/")
	if len(parts) < 3 || parts[0] != "product" {
		log.Printf("invalid topic format: %s", msg.Topic())
		return
	}

	stationMAC := parts[1]
	barcode := parts[2]

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var responseProduct domain.Product

	product, err := db.Fetch(ctx, barcode)
	if err != nil {
		log.Printf("failed to fetch product %s: %v", barcode, err)
		responseProduct = domain.Product{
			Barcode: barcode,
			Valid:   false,
		}
	} else {
		responseProduct = *product
		responseProduct.Valid = true
		responseProduct.Name = godiacritics.Normalize(responseProduct.Name)
		responseProduct.Price = godiacritics.Normalize(responseProduct.Price)
	}

	respJSON, err := json.Marshal(responseProduct)
	if err != nil {
		log.Printf("failed to serialize response: %v", err)
		return
	}

	responseTopic := fmt.Sprintf("product/%s", stationMAC)
	for i := 0; i < maxPublishRetries; i++ {
		token := client.Publish(responseTopic, 1, false, respJSON)
		if token.WaitTimeout(1*time.Second) && token.Error() == nil {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}

	log.Printf("CRITICAL: Failed to publish response for %s", barcode)
}
