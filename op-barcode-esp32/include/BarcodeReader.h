#pragma once

#include <SoftwareSerial.h>

typedef std::array<char, 30> Barcode;

class BarcodeReader
{
private:
    SoftwareSerial softwareSerial;
    std::array<char, 30> barcodeBuffer;
    int8_t txPin, rxPin;
    uint32_t baudRate;
public:
    BarcodeReader(uint32_t baudRate, int8_t txPin, int8_t rxPin);

    void init();

    bool dataPresent();

    void readUntilDelimiter(int8_t delimiter, Barcode& barcode);
};
