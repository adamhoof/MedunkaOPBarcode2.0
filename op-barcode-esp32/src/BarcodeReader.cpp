#include "BarcodeReader.h"


BarcodeReader::BarcodeReader(uint32_t baudRate, int8_t txPin, int8_t rxPin)
        : barcodeBuffer {}, txPin(txPin), rxPin(rxPin), baudRate(baudRate)
{}

void BarcodeReader::init()
{
    softwareSerial.begin(this->baudRate, SWSERIAL_8N1, this->rxPin, this->txPin);
}

bool BarcodeReader::dataPresent()
{
    return softwareSerial.available() > 0;
}

void BarcodeReader::readUntilDelimiter(int8_t delimiter, Barcode& barcode)
{
    uint8_t numBytes = softwareSerial.readBytesUntil(delimiter, barcode.data(), barcode.size());
    barcode[numBytes] = '\0';
}
