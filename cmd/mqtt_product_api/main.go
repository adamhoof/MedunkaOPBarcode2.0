package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/adamhoof/MedunkaOPBarcode2.0/internal/database"
	"github.com/adamhoof/MedunkaOPBarcode2.0/internal/utils"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	dbHandler, err := database.New("postgres")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if closeErr := dbHandler.Close(); closeErr != nil {
			log.Printf("failed to close database connection: %s", closeErr)
		}
	}()

	mqttClient := utils.CreateSecureMQTTClient()
	utils.ConnectOrFail(mqttClient)

	jobQueue := make(chan mqtt.Message, utils.GetEnvAsInt("APP_JOB_QUEUE_SIZE"))
	workerCount := utils.GetEnvAsInt("APP_WORKER_COUNT")
	workerTimeout := utils.GetEnvAsDuration("APP_DB_TIMEOUT")
	requestTopic := utils.GetEnvOrPanic("MQTT_TOPIC_REQUEST")

	for i := 0; i < workerCount; i++ {
		go processJobs(dbHandler, mqttClient, jobQueue, workerTimeout)
	}

	token := mqttClient.Subscribe(requestTopic, 1, func(client mqtt.Client, msg mqtt.Message) {
		select {
		case jobQueue <- msg:
		default:
			log.Printf("WARNING: System overloaded. Dropping request for topic: %s", msg.Topic())
		}
	})

	if token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	log.Printf("Service started with %d workers", workerCount)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	close(jobQueue)
}
