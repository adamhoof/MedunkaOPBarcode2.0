#include <Arduino.h>
#include <Adafruit_ILI9341.h>
#include <SPI.h>
#include <arial.h>
#include <SoftwareSerial.h>
#include "credentials.h"
#include <WifiController.h>
#include <OTAHandler.h>
#include "PubSubClient.h"
#include <ArduinoJson.h>
#include "ProductData.h"

WifiController wifiController = WifiController();
SoftwareSerial softwareSerial {};
WiFiClient wifiClient = WiFiClient();

PubSubClient mqttClient = PubSubClient();

const uint8_t displayCSPin = 32;
const uint8_t displayDCPin = 26;
const uint8_t displayRSTPin = 25;
//23 MOSI SDI, 19 MISO SDO, 18 SCK
Adafruit_ILI9341 display = Adafruit_ILI9341(displayCSPin, displayDCPin, displayRSTPin);

bool firmwareUpdateAwaiting = false;

void updateFirmware()
{
    OTAHandler::setEvents();
    OTAHandler::init();

    for (;;) {
        OTAHandler::maintainConnection();
    }
}

void maintainMQTTConnection(void* parameters)
{
    for (;;) {
        //the loop() function of mqttClient returns status of mqtt connection as a result
        if (mqttClient.loop()) {
            vTaskDelay(200 / portTICK_PERIOD_MS);
            continue;
        }
        //yes, when disconnected it does not have to be disconnected theoretically, but it sometimes sets the flag weirdly
        mqttClient.disconnect();
        mqttClient.connect(hostname);
        mqttClient.subscribe(hostname);
    }
}

char* mqttRequestBuffer = nullptr;

void packMQTTRequestInfo(char* barcodeAsCharArray)
{
    StaticJsonDocument<200> jsonDoc;
    jsonDoc["ClientTopic"] = hostname;
    jsonDoc["Barcode"] = barcodeAsCharArray;
    jsonDoc["IncludeDiacritics"] = false;
    ArduinoJson6194_F1::serializeJson(jsonDoc, mqttRequestBuffer, 200);
}

DeserializationError unpackMQTTResponse(const byte* response, ProductData* productData)
{
    StaticJsonDocument<350> productDataAsJson;
    DeserializationError error = deserializeJson(productDataAsJson, response);

    productData->name = productDataAsJson["Name"];
    productData->price = productDataAsJson["Price"];
    productData->stock = productDataAsJson["Stock"];
    productData->unitOfMeasure = productDataAsJson["UnitOfMeasure"];
    productData->unitOfMeasureKoef = productDataAsJson["UnitOfMeasureKoef"];
    return error;
}

void printErrorMessage(const char* error)
{
    display.fillScreen(ILI9341_BLACK);
    display.setCursor(0, 20);
    display.setTextSize(2);
    display.setTextColor(ILI9341_GREEN);
    display.printf("\nZkuste prosim\nznovu...\n");
    display.setTextSize(1);
    display.setTextColor(ILI9341_RED);
    display.printf("\nAdmin -> check if product exists in db...\n");

    if (error == nullptr) {
        return;
    }
    display.printf("Error message: %s", error);
}

void printProductData(ProductData* productData)
{
    if (strcmp(productData->name, "") == 0) {
        printErrorMessage("Name not found...");
        return;
    }

    display.fillScreen(ILI9341_BLACK);
    display.setTextColor(ILI9341_WHITE);
    display.setTextSize(1);
    display.setCursor(0, 20);
    display.printf("\n%s\n\n", productData->name);

    display.setTextColor(ILI9341_GREEN);
    display.setTextSize(2);
    display.printf("Cena: %.6g kc\n", productData->price);

    display.setTextColor(ILI9341_WHITE);

    if (strcmp(productData->unitOfMeasure, "") > 0) {
        display.setTextSize(1);
        display.printf("Cena za %s: %.6g kc\n\n",
                       productData->unitOfMeasure,
                       productData->price * productData->unitOfMeasureKoef);
    }

    display.setTextSize(1);
    display.printf("Stock: %s", productData->stock);
}

ProductData* productData = nullptr;
bool* receivedMessage = nullptr;
bool* finishedPrinting = nullptr;

void messageHandler(char* topic, const byte* payload, unsigned int length)
{
    if (length == 1) {
        firmwareUpdateAwaiting = true;
        return;
    }

    while (!finishedPrinting) {
        delay(10);
    }

    DeserializationError error = unpackMQTTResponse(payload, productData);

    if (error) {
        printErrorMessage(error.c_str());
        return;
    }
    *receivedMessage = true;
    *finishedPrinting = false;
}

void setup()
{
    mqttRequestBuffer = (char*) malloc(sizeof(char) * 200);
    productData = (ProductData*) malloc(sizeof(ProductData));
    receivedMessage = (bool*) malloc(sizeof(bool));
    finishedPrinting = (bool*) malloc(sizeof(bool));
    *receivedMessage = false;
    *finishedPrinting = false;

    display.begin();
    display.setFont(&ariblk9pt8b);
    display.setRotation(1);
    display.fillScreen(ILI9341_BLACK);

    softwareSerial.begin(9600, SWSERIAL_8N1, 13, 15);
    Serial.begin(115200);

    wifiController.setHostname(hostname).setSSID(wiFiSSID).setPassword(wiFiPassword);
    wifiController.connect();

    mqttClient.setServer(mqttServer, mqttPort);
    mqttClient.setClient(wifiClient);
    mqttClient.connect(hostname);
    mqttClient.subscribe(hostname);
    mqttClient.setCallback(messageHandler);

    xTaskCreatePinnedToCore(
            maintainMQTTConnection,
            "keepOTAAlive",
            5000,
            nullptr,
            1,
            nullptr,
            CONFIG_ARDUINO_RUNNING_CORE
    );
    Serial.println("Setup rdy");
}

void loop()
{
    if (receivedMessage) {
        printProductData(productData);
        *receivedMessage = false;
        *finishedPrinting = true;
    }
    // enter into if statement when one of conditions is true, otherwise we do not care about the inside
    if (!WiFi.isConnected() || firmwareUpdateAwaiting) {
        if (!firmwareUpdateAwaiting) {
            wifiController.maintainConnection();
        }
        //TODO send ip address
        updateFirmware();
    }

    if (softwareSerial.available() > 0) {
        uint8_t buffer[22];
        uint8_t size = softwareSerial.readBytesUntil(13, buffer, 22);
        char barcodeAsCharArray[size];
        memcpy(barcodeAsCharArray, buffer, size);
        for (int i = 0; i < size; ++i) {
            if (!isDigit(barcodeAsCharArray[i])) {
                return;
            }
        }
        Serial.println(barcodeAsCharArray);
        packMQTTRequestInfo(barcodeAsCharArray);

        bool successfulPublish = mqttClient.publish(publishTopic, mqttRequestBuffer, false);
        while (!successfulPublish) {
            delay(100);
            if (mqttClient.connected()) {
                successfulPublish = mqttClient.publish(publishTopic, mqttRequestBuffer, false);
            }
        }
    }
}
