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

bool receivedProductData = false;
bool finishedPrinting = true;
bool firmwareUpdateAwaiting = false;

const char* const firmwareUpdateTopic = "/firmware_update";
const char* const productDataRequestTopic = "/get_product_data";
const std::string productDataResponseTopic = std::string(clientName) + productDataRequestTopic;
const char* const lightCommandTopic = "/light";

ProductDataResponse productDataResponse;

void mqttMessageHandler(char* topic, const byte* payload, unsigned int length)
{
    if (strstr(topic, productDataRequestTopic) != nullptr) {
        while (!finishedPrinting) {
            delay(100);
        }
        deserializeProductDataResponse(payload, productDataResponse);
        receivedProductData = true;

    } else if (strstr(topic, firmwareUpdateTopic) != nullptr) {
        firmwareUpdateAwaiting = !firmwareUpdateAwaiting;

    } else if (strstr(topic, lightCommandTopic)) {
        LightCommandData lightCommandData {};
        deserializeLightCommand(payload, lightCommandData);
        lightCommandData.state == true ? barcodeReader.lightOn() : barcodeReader.lightOff();
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

void printProductData(Adafruit_ILI9341& disp, const ProductDataResponse& productData)
{
    clearDisplay(disp);
    disp.setCursor(0, 20);
    disp.setTextSize(1);
    disp.setTextColor(ILI9341_WHITE);
    display.printf("\n%s\n\n", productData.name.c_str());

    display.setTextSize(2);
    disp.setTextColor(ILI9341_GREEN);
    display.printf("Cena: %.6g kc\n", productData.price);

    display.setTextSize(1);
    display.setTextColor(ILI9341_WHITE);
    if (strcmp(productData.unitOfMeasure.c_str(), "") > 0) {
        display.printf("Cena za %s: %.6g kc\n\n",
                       productData.unitOfMeasure.c_str(),
                       productData.price * productData.unitOfMeasureKoef);
    }

    display.printf("Stock: %hu", productData.stock);
}

void setup()
{
    Serial.begin(115200);

    wifiConnectionHandler.setEventHandler(ARDUINO_EVENT_WIFI_STA_CONNECTED, WiFiConnectHandler);
    if (!wifiConnectionHandler.connect()) {
        ESP.restart();
    }
    wifiConnectionHandler.setEventHandler(ARDUINO_EVENT_WIFI_STA_DISCONNECTED, WiFiDisconnectHandler);

    barcodeReader.init();
    initDisplay(display, 3);

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
    if (!mqttClient.loop()) {
        mqttClient.disconnect();
        mqttClient.connect(clientName);
        std::string subscribeTopic = clientName + std::string("/+");
        mqttClient.subscribe(subscribeTopic.c_str());
        mqttClient.setCallback(mqttMessageHandler);
    }

    if (receivedProductData) {
        finishedPrinting = false;
        printProductData(display, productDataResponse);
        finishedPrinting = true;
        receivedProductData = false;
    }

    if (barcodeReader.dataPresent()) {
        Barcode barcode {};
        barcodeReader.readUntilDelimiter(DELIMITER, barcode);

        SerializedProductDataRequestBuffer requestBuffer;
        if (serializeProductDataRequest(
                ProductDataRequest {
                        .barcode = barcode.data(),
                        .responseTopic = productDataResponseTopic.c_str(),
                        .includeDiacritics = false},
                requestBuffer) == SERIALIZATION_FAILED) {
            printErrorMessage(display, "\nZkuste prosim\nznovu...\n");
        }
        if (!productDataRequestSuccessful(productDataRequestTopic, requestBuffer, 5, 100)) {
            printErrorMessage(display, "\nZkuste prosim\nznovu...\n");
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
