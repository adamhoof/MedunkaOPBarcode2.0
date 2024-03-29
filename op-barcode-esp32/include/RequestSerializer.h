#pragma once

#include <array>
#include <ArduinoJson.h>

typedef std::array<char, 200> SerializedProductDataRequestBuffer;

enum SerializationStatus {
    SERIALIZATION_OK, SERIALIZATION_FAILED
};

struct ProductDataRequest
{
    const char* barcode;
    const char* responseTopic;
    bool includeDiacritics;
};

SerializationStatus serializeProductDataRequest(const ProductDataRequest& requestToSerialize, SerializedProductDataRequestBuffer& serializedRequest);
