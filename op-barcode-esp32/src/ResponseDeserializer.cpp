#include "ResponseDeserializer.h"

DeserializationStatus deserializeProductDataResponse(const byte* payload, ProductDataResponse& productDataResponse)
{
    StaticJsonDocument<350> jsonResponse;
    DeserializationError error = deserializeJson(jsonResponse, payload);
    if (error != DeserializationError::Ok) {
        return DESERIALIZATION_FAILED;
    }

    productDataResponse.name = jsonResponse["name"].as<std::string>();
    productDataResponse.price = jsonResponse["price"];
    productDataResponse.stock = jsonResponse["stock"];
    productDataResponse.unitOfMeasure = jsonResponse["unitOfMeasure"].as<std::string>();
    productDataResponse.unitOfMeasureKoef = jsonResponse["unitOfMeasureCoef"];

    return DESERIALIZATION_OK;
}

DeserializationStatus deserializeLightCommand(const byte* input, LightCommandData& command)
{
    StaticJsonDocument<30> jsonResponse;
    DeserializationError error = deserializeJson(jsonResponse, input);
    if (error != DeserializationError::Ok) {
        return DESERIALIZATION_FAILED;
    }

    command.state = jsonResponse["state"].as<bool>();

    return DESERIALIZATION_OK;
}
