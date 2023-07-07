#include "DisplayController.h"


void initDisplay(Adafruit_ILI9341& display, uint8_t rotation)
{
    display.begin();
    display.setRotation(rotation);
    clearDisplay(display);
}

void printErrorMessage(Adafruit_ILI9341& display, const char* erroMessage)
{
    clearDisplay(display);
    display.setCursor(0, 20);
    display.setTextSize(2);
    display.setTextColor(ILI9341_RED);
    display.print(erroMessage);
}

void clearDisplay(Adafruit_ILI9341& display)
{
    display.fillScreen(ILI9341_BLACK);
}
