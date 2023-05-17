package api

import (
	"encoding/json"
	godiacritics "github.com/Regis24GmbH/go-diacritics"
	"github.com/adamhoof/MedunkaOPBarcode2.0/config"
	"github.com/adamhoof/MedunkaOPBarcode2.0/database"
	product_data "github.com/adamhoof/MedunkaOPBarcode2.0/product-data"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
	"time"
)

func GetProductData(dbHandler database.DatabaseHandler, conf *config.Config) mqtt.MessageHandler {
	return func(mqttClient mqtt.Client, message mqtt.Message) {
		var request product_data.ProductDataRequest
		err := json.Unmarshal(message.Payload(), &request)
		if err != nil {
			log.Println("error unpacking payload into product data request struct: ", err)
			return
		}

		productData, err := dbHandler.FetchProductData(conf.HTTPDatabaseUpdate.TableName, request.Barcode)
		if err != nil {
			log.Println("failed to fetch product data: ", err)
		}

		if !request.IncludeDiacritics {
			productData.Name = godiacritics.Normalize(productData.Name)
			productData.Price = godiacritics.Normalize(productData.Price)
		}

		productDataAsJson, err := json.Marshal(&productData)
		if err != nil {
			log.Println("unable to serialize product data into json: ", err)
			return
		}

		for {
			token := mqttClient.Publish(request.ClientTopic, 0, false, productDataAsJson)
			if token.WaitTimeout(5*time.Second) && token.Error() == nil {
				break
			}
			log.Println("failed to publish message...", token.Error())
			time.Sleep(1 * time.Second)
		}
	}
}
