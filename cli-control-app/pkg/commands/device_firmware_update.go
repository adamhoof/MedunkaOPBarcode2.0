package commands

func UpdateDeviceFirmware(deviceName string) {
	/*options := mqtt.NewClientOptions()
	options.AddBroker(os.Getenv("MQTT_SERVER_AND_PORT"))
	options.SetClientID("mqtt_test")
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
		log.Println("mqtt client failed to connect, retrying...", token.Error())
		time.Sleep(5 * time.Second)
	}

	log.Println("mqtt client connected")

	token := mqttClient.Subscribe("/test_topic", 0, func(client mqtt.Client, message mqtt.Message) {
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
		ClientTopic:       "/test_topic",
		Barcode:           "8595020340103",
		IncludeDiacritics: true,
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
	log.Println("mqtt client disconnected")*/
}
