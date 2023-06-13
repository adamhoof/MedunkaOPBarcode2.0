#include "RequestSerializer.h"

SerializationStatus serialize(const Request& requestToSerialize, SerializedRequestBuffer& serializedRequest)
{
    StaticJsonDocument<200> jsonDoc;
    jsonDoc["ClientTopic"] = requestToSerialize.responseTopic;
    jsonDoc["Barcode"] = requestToSerialize.barcode;
    jsonDoc["IncludeDiacritics"] = requestToSerialize.includeDiacritics;

    if (serializeJson(jsonDoc, serializedRequest.data(), serializedRequest.size()) == 0) {
        return SERIALIZATION_FAILED;
    }
    return SERIALIZATION_OK;
}
