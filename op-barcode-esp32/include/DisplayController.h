#pragma once

#include <Adafruit_ILI9341.h>

void initDisplay(Adafruit_ILI9341& display, uint8_t rotation);

void printErrorMessage(Adafruit_ILI9341& display, const char* erroMessage);

void clearDisplay(Adafruit_ILI9341& display);