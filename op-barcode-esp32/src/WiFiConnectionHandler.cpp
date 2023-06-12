#include "WiFiConnectionHandler.h"

WiFiConnectionHandler::WiFiConnectionHandler(const char* ssid, const char* password) : ssid(ssid), password(password)
{}

bool WiFiConnectionHandler::connect()
{
    if (!WiFi.isConnected()) {
        WiFi.begin(this->ssid.c_str(), this->password.c_str());
    }

    for (int i = 0; i < 25; ++i) {
        delay(200);
        if (WiFi.isConnected()){
            return true;
        }
    }
    return false;
}

void WiFiConnectionHandler::disconnect()
{
    WiFi.disconnect();
}

bool WiFiConnectionHandler::reconnect()
{
    this->disconnect();
    return this->connect();
}

void WiFiConnectionHandler::setEventHandler(arduino_event_id_t event, WiFiEventCb handler)
{
    WiFi.onEvent(handler, event);
}
