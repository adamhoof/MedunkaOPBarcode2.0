#include "BarcodeReader.h"


BarcodeReader::BarcodeReader(uint32_t baudRate, int8_t txPin, int8_t rxPin)
        : txPin(txPin), rxPin(rxPin), baudRate(baudRate)
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

void BarcodeReader::lightOn()
{
    softwareSerial.write(turnOnLight.data(), turnOnLight.size());
    delay(100);
    softwareSerial.write(saveToFlash.data(), saveToFlash.size());
    delay(200);
    softwareSerial.write(deepSleep.data(), deepSleep.size());
    delay(1000);
    softwareSerial.write(wakeUpModule.data(), wakeUpModule.size());
}

void BarcodeReader::lightOff()
{
    softwareSerial.write(turnOffLight.data(), turnOffLight.size());
    delay(100);
    softwareSerial.write(saveToFlash.data(), saveToFlash.size());
    delay(200);
    softwareSerial.write(deepSleep.data(), deepSleep.size());
    delay(1000);
    softwareSerial.write(wakeUpModule.data(), wakeUpModule.size());
}
