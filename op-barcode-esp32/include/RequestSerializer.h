#pragma once

#include <array>
#include <ArduinoJson.h>

typedef std::array<char, 200> SerializedRequestBuffer;

enum SerializationStatus {
    SERIALIZATION_OK, SERIALIZATION_FAILED
};

struct Request
{
    const char* barcode;
    const char* responseTopic;
    bool includeDiacritics;
};

SerializationStatus serialize(const Request& requestToSerialize, SerializedRequestBuffer& serializedRequest);
