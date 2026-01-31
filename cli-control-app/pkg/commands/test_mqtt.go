package commands

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/adamhoof/MedunkaOPBarcode2.0/internal/domain"
	"github.com/adamhoof/MedunkaOPBarcode2.0/mqtt-client"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func MQTTProductDataRequestTest(clientTopic string, barcode string, includeDiacritics bool) {
	mqttClient := mqtt_client.CreateDefault(clientTopic)
	mqtt_client.ConnectDefault(&mqttClient)

	token := mqttClient.Subscribe(clientTopic, 0, func(client mqtt.Client, message mqtt.Message) {
		log.Printf("Received request at topic: %s\n", message.Topic())
		var productData domain.Product
		err := json.Unmarshal(message.Payload(), &productData)
		if err != nil {
			log.Println("error unpacking payload into product data request struct: ", err)
			return
		}
		log.Println(productData)
	})
	if token.WaitTimeout(2*time.Second) && token.Error() != nil {
		log.Fatal("failed to subscribe: ", token.Error())
	}

	productDataRequest := domain.ProductDataRequest{
		ClientTopic:       clientTopic,
		Barcode:           barcode,
		IncludeDiacritics: includeDiacritics,
	}

	productDataAsJson, err := json.Marshal(&productDataRequest)
	if err != nil {
		log.Println("unable to serialize product data into json: ", err)
		return
	}

	for {
		token = mqttClient.Publish(os.Getenv("MQTT_PRODUCT_DATA_REQUEST"), 1, false, productDataAsJson)
		if token.WaitTimeout(5*time.Second) && token.Error() == nil {
			break
		}
		log.Println("failed to publish message...", token.Error())
		time.Sleep(1 * time.Second)
	}

	time.Sleep(time.Millisecond * 500)
	mqttClient.Disconnect(0)
	log.Println("mqtt client disconnected")
}
