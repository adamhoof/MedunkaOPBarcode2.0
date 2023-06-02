#pragma once

#include <string>

extern "C" {

#include "esp_wifi.h"
#include "esp_event.h"
#include "esp_log.h"
}

static const char *TAG = "WiFi Controller";

static void wifi_event_handler(void* arg, esp_event_base_t event_base,
                               int32_t event_id, void* event_data)
{
    if (event_id == WIFI_EVENT_STA_START) {
        ESP_LOGI(TAG, "starting wifi");
        esp_wifi_connect();
    } else if (event_id == WIFI_EVENT_STA_DISCONNECTED) {
        ESP_LOGW(TAG, "wifi disconnected, reconnecting...");
        esp_wifi_connect();
    } else if (event_id == IP_EVENT_STA_GOT_IP) {
        auto event = (ip_event_got_ip_t*) event_data;
        ESP_LOGI(TAG, "got ip: " IPSTR, IP2STR(&event->ip_info.ip));
    }
}

class WiFiController
{
public:
    WiFiController(const wifi_config_t& conf);
    /*void set_config(const wifi_config_t& conf);*/
    void start();

private:
    void load_config();
    static void wifi_event_handler(void* arg, esp_event_base_t event_base,
                                   int32_t event_id, void* event_data);
    void register_event_handlers();
    wifi_config_t wifi_config;
};


