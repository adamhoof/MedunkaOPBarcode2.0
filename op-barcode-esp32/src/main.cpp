#include <Arduino.h>
#include "WiFiConnectionHandler.h"
#include "BarcodeReader.h"
#include "credentials.h"
#include <ArduinoJson.h>
#include <SPI.h>
#include <Adafruit_ILI9341.h>
#include "DisplayController.h"
#include "RequestSerializer.h"
#include "ResponseDeserializer.h"
#include <PubSubClient.h>

#define TX 15
#define RX 13
#define BAUD 9600
#define DELIMITER '\r'

#define CS  32
#define DC  26
#define RST  25
//23 MOSI/SDI, 19 MISO/SDO, 18 SCK, set implicitly by library

WiFiConnectionHandler wifiConnectionHandler(clientName, ssid, password);
BarcodeReader barcodeReader(BAUD, TX, RX);
WiFiClient wiFiClient;
PubSubClient mqttClient;
Adafruit_ILI9341 display(CS, DC, RST);
DisplayController displayController(display);


bool receivedMessage = false;
bool finishedPrinting = true;

Response* response;

void messageHandler(char* topic, const byte* payload, unsigned int length)
{
    while (!finishedPrinting) {
        delay(100);
    }

    deserializeResponse(payload, response);

    receivedMessage = true;
}

static void WiFiDisconnectHandler(arduino_event_id_t eventId)
{
    Serial.println("Disconnected, connecting bacc");
    if (!wifiConnectionHandler.reconnect()) {
        ESP.restart();
    }
}

static void WiFiConnectHandler(arduino_event_id_t eventId)
{
    Serial.printf("Connected to %s!\n", ssid);
}


void setup()
{
    Serial.begin(115200);
    response = new Response;

    wifiConnectionHandler.setEventHandler(ARDUINO_EVENT_WIFI_STA_CONNECTED, WiFiConnectHandler);
    if (!wifiConnectionHandler.connect()) {
        ESP.restart();
    }
    wifiConnectionHandler.setEventHandler(ARDUINO_EVENT_WIFI_STA_DISCONNECTED, WiFiDisconnectHandler);

    barcodeReader.init();

    mqttClient.setServer(mqttServer, mqttPort);
    mqttClient.setClient(wiFiClient);
    mqttClient.connect(clientName);
    std::string subscribeTopic = clientName + std::string("/+");
    mqttClient.subscribe(subscribeTopic.c_str());
    mqttClient.setCallback(messageHandler);
}

void loop()
{
     mqttClient.loop();

     if (receivedMessage) {
         //print
     }

     if (barcodeReader.dataPresent()) {
         Barcode barcode {};
         barcodeReader.readUntilDelimiter(DELIMITER, barcode);

         for (const char& charie: barcode) {
             Serial.println(charie);
         }
         SerializedRequestBuffer buffer;
         if (serialize(Request {.barcode = barcode.data(), .responseTopic = (std::string (clientName) + "/get_product_data").c_str(), .includeDiacritics = false},
                       buffer) == SERIALIZATION_FAILED) {
             //pee pee poo poo
         }
     }
}
