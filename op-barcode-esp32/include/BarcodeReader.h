#pragma once

#include <SoftwareSerial.h>

typedef std::array<char, 30> Barcode;

class BarcodeReader
{
private:
    const std::array<byte, 9> turnOnLight = {0x7E, 0x00, 0x08, 0x01, 0x00, 0x00, 0x5F, 0xAB, 0xCD};
    const std::array<byte, 9> turnOffLight = {0x7E, 0x00, 0x08, 0x01, 0x00, 0x00, 0x53, 0xAB, 0xCD};
    const std::array<byte, 9> saveToFlash = {0x7E, 0x00, 0x09, 0x01, 0x00, 0x00, 0x00, 0xDE, 0xC8};
    const std::array<byte, 9> deepSleep = {0x7E, 0x00, 0x08, 0x01, 0x00, 0xD9, 0xA5, 0xAB, 0xCD};
    const std::array<byte, 9> wakeUpModule = {0x7E, 0x00, 0x08, 0x01, 0x00, 0xD9, 0x00, 0xAB, 0xCD};

    int8_t txPin, rxPin;
    uint32_t baudRate;
public:
    SoftwareSerial softwareSerial;

    BarcodeReader(uint32_t baudRate, int8_t txPin, int8_t rxPin);

    void init();

    bool dataPresent();

    void readUntilDelimiter(int8_t delimiter, Barcode& barcode);

    void lightOn();

    void lightOff();
};
