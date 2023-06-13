#pragma once

#include <Arduino.h>
#include <ArduinoJson.h>

enum DeserializationStatus {
    DESERIALIZATION_OK, DESERIALIZATION_FAILED
};

struct Response {
    const char* name;
    double price;
    uint16_t stock;
    const char* unitOfMeasure;
    double unitOfMeasureKoef;
};

DeserializationStatus deserializeResponse(const byte* buffer, Response* const response);
