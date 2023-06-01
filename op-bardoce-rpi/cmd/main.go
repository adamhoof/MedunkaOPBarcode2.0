package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/adamhoof/MedunkaOPBarcode2.0/op-bardoce-rpi/pkg/cli_artist"
	"github.com/adamhoof/MedunkaOPBarcode2.0/op-bardoce-rpi/pkg/mqtt/mqtt_handlers"
	product_data "github.com/adamhoof/MedunkaOPBarcode2.0/product-data"
	mqtt "github.com/eclipse/paho.mqtt.golang"
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

	baud, err := strconv.Atoi(os.Getenv(""))
	if err != nil {
		log.Fatalf("failed to convert %s to number: %s", os.Getenv(""), err)
	}

	serialPort, err := serial.OpenPort(&serial.Config{Name: os.Getenv(""), Baud: baud})
	if err != nil {
		log.Fatalf("failed to initialize barcode reader: %s", err)
	}
	barcodeReader := bufio.NewReader(serialPort)
	log.Println("serial port initialized")

	options := mqtt.NewClientOptions()
	options.AddBroker(os.Getenv("MQTT_SERVER_AND_PORT"))
	options.SetClientID(os.Args[2])
	options.SetAutoReconnect(true)
	options.SetConnectRetry(true)
	options.SetCleanSession(false)
	options.SetConnectRetryInterval(time.Second * 2)
	options.SetOrderMatters(false)

	mqttClient := mqtt.NewClient(options)

	for {
		token := mqttClient.Connect()
		if token.WaitTimeout(5*time.Second) && token.Error() == nil {
			break
		}
		log.Println("handlers client failed to connect, retrying...", token.Error())
		time.Sleep(5 * time.Second)

	}
	log.Println("handlers client connected")

	mqttClient.Subscribe(os.Args[2], 0, mqtt_handlers.ProductDataResponseHandler())

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
			ClientTopic:       os.Args[2],
			Barcode:           barcodeAsString,
			IncludeDiacritics: true,
		}

		jsonProductDataRequest, err := json.Marshal(productDataRequest)
		if err != nil {
			fmt.Println("failed to pack data into productDataRequest")
		}

		for {
			token := mqttClient.Publish(os.Getenv("MQTT_PRODUCT_DATA_REQUEST"), 0, false, jsonProductDataRequest)
			if token.WaitTimeout(5*time.Second) && token.Error() == nil {
				break
			}
			log.Println("failed to publish message...", token.Error())
			time.Sleep(1 * time.Second)
		}
	}
}
