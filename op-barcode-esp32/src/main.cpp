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
#include <OTAHandler.h>

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


bool receivedProductData = false;
bool finishedPrinting = true;
bool firmwareUpdateAwaiting = false;

const char* const firmwareUpdateTopic = "/firmware_update";
const char* const productDataRequestTopic = "/get_product_data";
const std::string productDataResponseTopic = std::string(clientName) + productDataRequestTopic;
const char* const lightCommandTopic = "/light";

ProductDataResponse* response;

void mqttMessageHandler(char* topic, const byte* payload, unsigned int length)
{
    if (strstr(topic, productDataRequestTopic) != nullptr) {
        while (!finishedPrinting) {
            delay(100);
        }
        deserializeProductDataResponse(payload, response);
        receivedProductData = true;

    } else if (strstr(topic, firmwareUpdateTopic) != nullptr) {
        firmwareUpdateAwaiting = !firmwareUpdateAwaiting;

    } else if (strstr(topic, lightCommandTopic)) {
        bool requestedLightStatus = true;
        deserializeLightCommand(payload, requestedLightStatus);
        requestedLightStatus == true ? barcodeReader.lightOn() : barcodeReader.lightOff();
    } else {

    }
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

bool productDataRequestSuccessful(const char* const requestTopic, const SerializedProductDataRequestBuffer& buffer,
                                  uint8_t maxPublishRetries, uint32_t publishDelay)
{
    uint8_t publishRetries = 0;
    while (publishRetries < maxPublishRetries) {
        if (mqttClient.publish(requestTopic, buffer.data(), false)) {
            break;
        }
        delay(publishDelay);
        publishRetries++;
    }

    if (publishRetries == maxPublishRetries) {
        return false;
    }
    return true;
}

void setup()
{
    Serial.begin(115200);
    response = new ProductDataResponse;

    wifiConnectionHandler.setEventHandler(ARDUINO_EVENT_WIFI_STA_CONNECTED, WiFiConnectHandler);
    if (!wifiConnectionHandler.connect()) {
        ESP.restart();
    }
    wifiConnectionHandler.setEventHandler(ARDUINO_EVENT_WIFI_STA_DISCONNECTED, WiFiDisconnectHandler);

    /*barcodeReader.init();*/

    mqttClient.setServer(mqttServer, mqttPort);
    mqttClient.setClient(wiFiClient);
    mqttClient.connect(clientName);
    std::string subscribeTopic = clientName + std::string("/+");
    mqttClient.subscribe(subscribeTopic.c_str());
    mqttClient.setCallback(mqttMessageHandler);
}

void loop()
{
    while (!WiFi.isConnected()) {
        delay(10);
    }
    mqttClient.loop();

    if (receivedProductData) {
        // print to display
    }

    if (barcodeReader.dataPresent()) {
        Barcode barcode {};
        barcodeReader.readUntilDelimiter(DELIMITER, barcode);

        for (const char& digit: barcode) {
            if (!isDigit(digit)) {
                return;
                // print to display
            }
        }
        SerializedProductDataRequestBuffer requestBuffer;
        if (serializeProductDataRequest(
                ProductDataRequest {
                        .barcode = barcode.data(),
                        .responseTopic = productDataResponseTopic.c_str(),
                        .includeDiacritics = false},
                requestBuffer) == SERIALIZATION_FAILED) {
            // serialization failed, print to display
        }
        if (!productDataRequestSuccessful(productDataRequestTopic, requestBuffer, 5, 100)) {
            // print to display
        }
    }

    if (firmwareUpdateAwaiting) {
        OTAHandler::init();
        OTAHandler::setEvents();
        while (firmwareUpdateAwaiting) {
            OTAHandler::maintainConnection();
        }
    }
}
