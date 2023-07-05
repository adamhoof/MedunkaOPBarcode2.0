#include "RequestSerializer.h"

SerializationStatus serializeProductDataRequest(const ProductDataRequest& requestToSerialize, SerializedProductDataRequestBuffer& serializedRequest)
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
