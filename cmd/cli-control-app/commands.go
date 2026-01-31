package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/adamhoof/MedunkaOPBarcode2.0/internal/domain"
	"github.com/adamhoof/MedunkaOPBarcode2.0/internal/mqtt-client"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type ResponseContent struct {
	Message string `json:"message"`
}

func DatabaseUpdate() error {
	if err := fileReadyToBeUsed(os.Getenv("MDB_PATH")); err != nil {
		return fmt.Errorf("file not ready to be used: %w", err)
	}

	if err := sendFileToServer(os.Getenv("HTTP_SERVER_HOST"), os.Getenv("HTTP_SERVER_PORT"), os.Getenv("HTTP_SERVER_UPDATE_ENDPOINT"), os.Getenv("MDB_PATH")); err != nil {
		return fmt.Errorf("failed to send file to server: %w", err)
	}
	return nil
}

func fileReadyToBeUsed(filePath string) error {
	_, err := os.Stat(filePath)
	if err == nil {
		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("permission denied: cannot read file at %s", filePath)
		}
		if err := file.Close(); err != nil {
			log.Println("failed to close file")
		}
		return nil
	}
	if os.IsNotExist(err) {
		return fmt.Errorf("file not found at the specified path: %s", filePath)
	}
	return err
}

func sendFileToServer(host string, port string, endpoint string, fileLocation string) error {
	file, err := os.Open(fileLocation)
	if err != nil {
		return err
	}
	defer func() {
		err = file.Close()
		if err != nil {
			log.Printf("failed to close file: %s\n", err)
		}
	}()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(fileLocation))
	if err != nil {
		return err
	}
	if _, err = io.Copy(part, file); err != nil {
		return err
	}

	err = writer.Close()
	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://%s:%s%s", host, port, endpoint)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to do request: %s", err)
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			log.Printf("failed to close response body: %s\n", err)
		}
	}()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %s", err)
	}

	responseContent := ResponseContent{}
	if err = json.Unmarshal(bodyBytes, &responseContent); err != nil {
		return fmt.Errorf("failed to parse server response: %s", err)
	}

	fmt.Printf("%s\n%s\n", responseContent.Message, resp.Status)
	return nil
}

type LightCommand struct {
	State bool `json:"state"`
}

func TurnOnLight(topic string) {
	mqttClient := mqtt_client.CreateDefault("light_controller")
	mqtt_client.ConnectDefault(&mqttClient)

	lightCommand := LightCommand{State: true}
	lightCommandAsJSON, err := json.Marshal(&lightCommand)
	if err != nil {
		log.Println("unable to serialize light command into json: ", err)
		return
	}

	for {
		token := mqttClient.Publish(topic, 1, false, lightCommandAsJSON)
		if token.WaitTimeout(5*time.Second) && token.Error() == nil {
			break
		}
		log.Println("failed to publish message...", token.Error())
		time.Sleep(1 * time.Second)
	}
	mqttClient.Disconnect(0)
	log.Println("mqtt client disconnected")
}

func TurnOffLight(topic string) {
	mqttClient := mqtt_client.CreateDefault("light_controller")
	mqtt_client.ConnectDefault(&mqttClient)

	lightCommand := LightCommand{State: false}
	lightCommandAsJSON, err := json.Marshal(&lightCommand)
	if err != nil {
		log.Println("unable to serialize light command into json: ", err)
		return
	}

	for {
		token := mqttClient.Publish(topic, 1, false, lightCommandAsJSON)
		if token.WaitTimeout(5*time.Second) && token.Error() == nil {
			break
		}
		log.Println("failed to publish message...", token.Error())
		time.Sleep(1 * time.Second)
	}
	mqttClient.Disconnect(0)
	log.Println("mqtt client disconnected")
}

func UpdateFirmware(topic string) {
	mqttClient := mqtt_client.CreateDefault("firmware_updater")
	mqtt_client.ConnectDefault(&mqttClient)

	for {
		token := mqttClient.Publish(topic, 1, false, "true")
		if token.WaitTimeout(5*time.Second) && token.Error() == nil {
			break
		}
		log.Println("failed to publish message...", token.Error())
		time.Sleep(1 * time.Second)
	}
	mqttClient.Disconnect(0)
	log.Println("mqtt client disconnected")
}

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
