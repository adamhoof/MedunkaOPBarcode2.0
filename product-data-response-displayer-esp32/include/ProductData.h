#pragma once

struct ProductData
{
public:
    const char* name {};
    float price {};
    const char* stock {};
    const char* unitOfMeasure {};
    float unitOfMeasureKoef {};

    ProductData();
};
