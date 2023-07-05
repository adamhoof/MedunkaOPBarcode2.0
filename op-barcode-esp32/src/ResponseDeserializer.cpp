#include "ResponseDeserializer.h"

DeserializationStatus deserializeProductDataResponse(const byte* const input, ProductDataResponse* const response)
{
    StaticJsonDocument<350> jsonResponse;
    response->name = jsonResponse["Name"];
    response->price = jsonResponse["Price"];
    response->stock = jsonResponse["Stock"];
    response->unitOfMeasure = jsonResponse["UnitOfMeasure"];
    response->unitOfMeasureKoef = jsonResponse["UnitOfMeasureKoef"];
    DeserializationError error = deserializeJson(jsonResponse, input);
    if (error != DeserializationError::Ok) {
        return DESERIALIZATION_FAILED;
    }
    return DESERIALIZATION_OK;
}

DeserializationStatus deserializeLightCommand(const byte* const input, bool& state)
{
    const char* payloadStr = reinterpret_cast<const char*>(input);

    if (strcmp(payloadStr, "true") == 0 || strcmp(payloadStr, "1") == 0) {
        state = true;
        return DESERIALIZATION_OK;
    } else if (strcmp(payloadStr, "false") == 0 || strcmp(payloadStr, "0") == 0) {
        state = false;
        return DESERIALIZATION_OK;
    }
    return DESERIALIZATION_FAILED;
}
