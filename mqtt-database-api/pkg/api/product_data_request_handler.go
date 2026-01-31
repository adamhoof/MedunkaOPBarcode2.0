package api

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
	requestTimeout = 5 * time.Second
)

func HandleProductDataRequest(dbHandler database.Handler) mqtt.MessageHandler {
	return func(mqttClient mqtt.Client, message mqtt.Message) {
		var request domain.ProductDataRequest
		if err := json.Unmarshal(message.Payload(), &request); err != nil {
			log.Printf("error unpacking payload into product data request struct: %s", err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
		defer cancel()

		productData, err := dbHandler.Fetch(ctx, request.Barcode)
		if err != nil {
			log.Printf("failed to fetch product data: %s", err)
			return
		}

		if !request.IncludeDiacritics {
			productData.Name = godiacritics.Normalize(productData.Name)
			productData.Price = godiacritics.Normalize(productData.Price)
		}

		productDataAsJSON, err := json.Marshal(productData)
		if err != nil {
			log.Printf("unable to serialize product data into json: %s", err)
			return
		}

		for {
			token := mqttClient.Publish(request.ClientTopic, 1, false, productDataAsJSON)
			if token.WaitTimeout(5*time.Second) && token.Error() == nil {
				break
			}
			log.Printf("failed to publish message: %s", token.Error())
			time.Sleep(1 * time.Second)
		}
	}
}
