#include <Arduino.h>
#include "WiFiConnectionHandler.h"
#include "BarcodeReader.h"
#include "credentials.h"
#include <ArduinoJson.h>
#include <SPI.h>
#include <Adafruit_ILI9341.h>
#include "RequestSerializer.h"

#define TX 15
#define RX 13
#define BAUD 9600
#define DELIMITER '\r'

WiFiConnectionHandler wifiConnectionHandler(ssid, password);
BarcodeReader barcodeReader(BAUD, TX, RX);

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

    wifiConnectionHandler.setEventHandler(ARDUINO_EVENT_WIFI_STA_CONNECTED, WiFiConnectHandler);
    if (!wifiConnectionHandler.connect()) {
        ESP.restart();
    }
    wifiConnectionHandler.setEventHandler(ARDUINO_EVENT_WIFI_STA_DISCONNECTED, WiFiDisconnectHandler);

    barcodeReader.init();
}

void loop()
{
    if (barcodeReader.dataPresent()) {
        Barcode barcode {};
        barcodeReader.readUntilDelimiter(DELIMITER, barcode);

    }
}
