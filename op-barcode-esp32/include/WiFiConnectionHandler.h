#pragma once

#include "WiFi.h"
#include "esp_wifi.h"

class WiFiConnectionHandler {
private:
    std::string ssid;
    std::string password;

public:
    WiFiConnectionHandler(const char* ssid, const char* password);

    bool connect();
    void disconnect();
    bool reconnect();
    void setEventHandler(arduino_event_id_t event, WiFiEventCb handler);
};
