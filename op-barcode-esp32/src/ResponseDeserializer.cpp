#include "ResponseDeserializer.h"

DeserializationStatus deserializeProductDataResponse(const byte* const input, ProductDataResponse& response)
{
    StaticJsonDocument<350> jsonResponse;
    DeserializationError error = deserializeJson(jsonResponse, input);

    response.name = jsonResponse["name"].as<const char*>();
    response.price = jsonResponse["price"].as<double>();
    response.stock = jsonResponse["stock"].as<uint16_t>();
    response.unitOfMeasure = jsonResponse["unitOfMeasure"].as<const char*>();
    response.unitOfMeasureKoef = jsonResponse["unitOfMeasureCoef"].as<double>();
    if (error != DeserializationError::Ok) {
        return DESERIALIZATION_FAILED;
    }
    return DESERIALIZATION_OK;
}

DeserializationStatus deserializeLightCommand(const byte* const input, LightCommandData& command)
{
    StaticJsonDocument<30> jsonResponse;
    DeserializationError error = deserializeJson(jsonResponse, input);

    if (error != DeserializationError::Ok) {
        return DESERIALIZATION_FAILED;
    }

    command.state = jsonResponse["state"].as<bool>();

    return DESERIALIZATION_OK;
}