#include <Arduino.h>
#include "WiFiConnectionHandler.h"
#include "credentials.h"
#include <SoftwareSerial.h>
#include <SPI.h>
#include <Adafruit_ILI9341.h>

WiFiConnectionHandler wifiConnectionHandler = WiFiConnectionHandler(ssid, password);

static void WiFiDisconnectHandler(arduino_event_id_t eventId)
{
    Serial.println("Disconnected, connecting bacc");
    if (!wifiConnectionHandler.reconnect()){
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
}

void loop()
{

}
