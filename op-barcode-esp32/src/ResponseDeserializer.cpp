#include "ResponseDeserializer.h"

DeserializationStatus deserializeResponse(const byte* const input, Response* const response)
{
    StaticJsonDocument<350> jsonResponse;
    response->name = jsonResponse["Name"];
    response->price = jsonResponse["Price"];
    response->stock = jsonResponse["Stock"];
    response->unitOfMeasure = jsonResponse["UnitOfMeasure"];
    response->unitOfMeasureKoef = jsonResponse["UnitOfMeasureKoef"];
    DeserializationError error = deserializeJson(jsonResponse, input);
    if (error != DeserializationError::Ok ) {
        return DESERIALIZATION_FAILED;
    }
    return DESERIALIZATION_OK;
}

