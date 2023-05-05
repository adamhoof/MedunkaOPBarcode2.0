package main

import (
	barcode "MedunkaOpBarcodeMQTT/OpBarcodeRPI/pkg/Barcode"
	cli_artist "MedunkaOpBarcodeMQTT/OpBarcodeRPI/pkg/CLIArtist"
	essential "MedunkaOpBarcodeMQTT/OpBarcodeRPI/pkg/EssentialConfig"
	response_handler "MedunkaOpBarcodeMQTT/OpBarcodeRPI/pkg/GetProductDataResponseHandler"
	serial_communication "MedunkaOpBarcodeMQTT/OpBarcodeRPI/pkg/SerialCommunication"
	product_data "MedunkaOpBarcodeMQTT/ProductData"
	"encoding/json"
	"fmt"
	typeconv "github.com/adamhoof/GolangTypeConvertorWrapper/pkg"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/tarm/serial"
	"os"
)

func main() {
	var conf essential.Config
	essential.LoadConfigFromJsonFile(os.Args[1], &conf)
	conf.MQTTClientName = "OpBarcodeRPI"

	serialPort, serialErr := serial_communication.OpenPort(&serial.Config{Name: conf.SerialPortName, Baud: conf.SerialPortBaudRate})
	if serialErr != nil {
		fmt.Println(serialErr)
		return
	}

	var barcodeReaderHandler barcode.ReaderHandler
	barcodeReaderHandler.GetPort(serialPort)

	options := mqtt.ClientOptions{}
	options.AddBroker(conf.MQTTServer)
	options.SetClientID(conf.MQTTClientName)
	options.SetAutoReconnect(true)
	options.SetConnectRetry(true)
	options.SetCleanSession(false)
	options.SetOrderMatters(false)

	mqttClient := mqtt.NewClient(&options)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	fmt.Println("MQTT client connection established?: ", mqttClient.IsConnected())

	mqttClient.Subscribe(fmt.Sprintf("/%s", conf.MQTTClientName), 0, response_handler.GetProductDataResponse())

	fmt.Println("rdy to scan boi")

	for {
		barcodeAsByteArray := barcodeReaderHandler.ReadUntilDelimiter(conf.BarcodeReadingTerminationDelimiter)
		cli_artist.ClearTerminal()

		barcodeAsString := typeconv.ByteArrayToString(barcodeAsByteArray[:len(barcodeAsByteArray)-1]) //cut out the delimiter and convert to string

		request := product_data.Request{
			ClientTopic:       fmt.Sprintf("/%s", conf.MQTTClientName),
			Barcode:           barcodeAsString,
			IncludeDiacritics: true,
		}

		requestAsJson, err := json.Marshal(request)
		if err != nil {
			fmt.Println("failed to pack data into request")
		}

		mqttClient.Publish("/get_product_data", 0, false, requestAsJson)
	}
}
