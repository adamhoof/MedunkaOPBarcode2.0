package commands

import (
	"encoding/json"
	mqtt_client "github.com/adamhoof/MedunkaOPBarcode2.0/mqtt-client"
	product_data "github.com/adamhoof/MedunkaOPBarcode2.0/product-data"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
	"os"
	"time"
)

func MQTTProductDataRequestTest(clientTopic string, barcode string, includeDiacritics bool) {
	mqttClient := mqtt_client.CreateDefault(clientTopic)
	mqtt_client.ConnectDefault(&mqttClient)

	token := mqttClient.Subscribe(clientTopic, 0, func(client mqtt.Client, message mqtt.Message) {
		log.Printf("Received request at topic: %s\n", message.Topic())
		var productData product_data.ProductData
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

	productData := product_data.ProductDataRequest{
		ClientTopic:       clientTopic,
		Barcode:           barcode,
		IncludeDiacritics: includeDiacritics,
	}

	productDataAsJson, err := json.Marshal(&productData)
	if err != nil {
		log.Println("unable to serialize product data into json: ", err)
		return
	}

	for {
		token = mqttClient.Publish(os.Getenv("MQTT_PRODUCT_DATA_REQUEST"), 0, false, productDataAsJson)
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
