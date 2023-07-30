package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	mqtt_client "github.com/adamhoof/MedunkaOPBarcode2.0/mqtt-client"
	"github.com/adamhoof/MedunkaOPBarcode2.0/op-barcode-rpi/pkg/cli_artist"
	"github.com/adamhoof/MedunkaOPBarcode2.0/op-barcode-rpi/pkg/mqtt/mqtt_handlers"
	product_data "github.com/adamhoof/MedunkaOPBarcode2.0/product-data"
	"github.com/joho/godotenv"
	"github.com/tarm/serial"
	"log"
	"os"
	"strconv"
	"time"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("required number of arguments: %d, provided: %d\nusage: executable env_file_location name_of_device", 3, len(os.Args))
	}
	err := godotenv.Load(os.Args[1])
	if err != nil {
		log.Fatalf("failed to load environment variables from file %s: %s", os.Args[1], err)
	}

	clientName := os.Args[2]
	productDataResponseTopic := clientName + "/" + os.Getenv("MQTT_PRODUCT_DATA_REQUEST")
	lightControlTopic := clientName + "/" + os.Getenv("LIGHT_CONTROL_TOPIC")

	baud, err := strconv.Atoi(os.Getenv("SERIAL_BAUD_RATE"))
	if err != nil {
		log.Fatalf("failed to convert %s to number: %s", os.Getenv("SERIAL_BAUD_RATE"), err)
	}

	serialPort, err := serial.OpenPort(&serial.Config{Name: os.Getenv("SERIAL_PORT_NAME"), Baud: baud})
	if err != nil {
		log.Fatalf("failed to initialize barcode reader: %s", err)
	}
	barcodeReader := bufio.NewReader(serialPort)
	log.Println("serial port initialized")

	mqttClient := mqtt_client.CreateDefault(clientName)
	mqtt_client.ConnectDefault(&mqttClient)

	log.Println("mqtt client connected")

	mqttClient.Subscribe(productDataResponseTopic, 1, mqtt_handlers.ProductDataResponseHandler())
	mqttClient.Subscribe(lightControlTopic, 1, mqtt_handlers.LightControlHandler(serialPort))

	var terminator byte = '\r'
	fmt.Println("rdy to scan boi")

	for {
		barcodeAsByteArray, err := barcodeReader.ReadBytes(terminator)
		cli_artist.ClearTerminal()

		if err != nil {
			log.Printf("failed to read barcode: %s\nplease try again...", err)
			continue
		}

		barcodeAsString := string(barcodeAsByteArray[:len(barcodeAsByteArray)-1])

		productDataRequest := product_data.ProductDataRequest{
			ClientTopic:       productDataResponseTopic,
			Barcode:           barcodeAsString,
			IncludeDiacritics: true,
		}

		jsonProductDataRequest, err := json.Marshal(productDataRequest)
		if err != nil {
			fmt.Println("failed to pack data into productDataRequest")
		}

		for {
			token := mqttClient.Publish(os.Getenv("MQTT_PRODUCT_DATA_REQUEST"), 1, false, jsonProductDataRequest)
			if token.WaitTimeout(5*time.Second) && token.Error() == nil {
				break
			}
			log.Println("failed to publish message...", token.Error())
			time.Sleep(1 * time.Second)
		}
	}
}
