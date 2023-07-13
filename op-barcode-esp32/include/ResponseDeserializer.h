#pragma once

#include <Arduino.h>
#include <ArduinoJson.h>

enum DeserializationStatus {
    DESERIALIZATION_OK, DESERIALIZATION_FAILED
};

struct ProductDataResponse {
    std::string name;
    double price;
    uint16_t stock;
    std::string unitOfMeasure;
    double unitOfMeasureKoef;
};

struct LightCommandData{
    bool state;
};

DeserializationStatus deserializeProductDataResponse(const byte* payload, ProductDataResponse& productDataResponse);

DeserializationStatus deserializeLightCommand(const byte* input, LightCommandData& command);
