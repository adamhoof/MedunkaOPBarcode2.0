#pragma once

#include <Arduino.h>
#include <ArduinoJson.h>

enum DeserializationStatus {
    DESERIALIZATION_OK, DESERIALIZATION_FAILED
};

struct ProductDataResponse {
    const char* name;
    double price;
    uint16_t stock;
    const char* unitOfMeasure;
    double unitOfMeasureKoef;
};

DeserializationStatus deserializeProductDataResponse(const byte* const input, ProductDataResponse* const response);
