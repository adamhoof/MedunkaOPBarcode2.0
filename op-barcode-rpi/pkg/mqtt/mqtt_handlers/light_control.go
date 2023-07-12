package mqtt_handlers

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/tarm/serial"
	"log"
	"time"
)

func sendLightOnCommand(port *serial.Port) error {
	turnOnLight := []byte{0x7E, 0x00, 0x08, 0x01, 0x00, 0x00, 0x5F, 0xAB, 0xCD}
	_, err := port.Write(turnOnLight)
	time.Sleep(100 * time.Millisecond)

	return err
}

func sendLightOffCommand(port *serial.Port) error {
	turnOffLight := []byte{0x7E, 0x00, 0x08, 0x01, 0x00, 0x00, 0x53, 0xAB, 0xCD}
	_, err := port.Write(turnOffLight)

	return err
}

func sendSaveToFlashCommand(port *serial.Port) error {
	saveToFlash := []byte{0x7E, 0x00, 0x09, 0x01, 0x00, 0x00, 0x00, 0xDE, 0xC8}
	_, err := port.Write(saveToFlash)

	return err
}

func sendDeepSleepCommand(port *serial.Port) error {
	deepSleep := []byte{0x7E, 0x00, 0x08, 0x01, 0x00, 0xD9, 0xA5, 0xAB, 0xCD}
	_, err := port.Write(deepSleep)

	return err
}

func sendWakeUpModuleCommand(port *serial.Port) error {
	wakeUpModule := []byte{0x7E, 0x00, 0x08, 0x01, 0x00, 0xD9, 0x00, 0xAB, 0xCD}
	_, err := port.Write(wakeUpModule)

	return err
}

func lightOn(port *serial.Port) {
	var err error
	if err = sendLightOnCommand(port); err == nil {
		time.Sleep(100 * time.Millisecond)
		if err = sendSaveToFlashCommand(port); err == nil {
			time.Sleep(200 * time.Millisecond)
			if err = sendDeepSleepCommand(port); err == nil {
				time.Sleep(1000 * time.Millisecond)
				if err = sendWakeUpModuleCommand(port); err == nil {
					return
				}
			}
		}
	}

	log.Printf("failed to write command: %s\n", err)
}

func lightOff(port *serial.Port) {
	var err error
	if err = sendLightOffCommand(port); err == nil {
		time.Sleep(100 * time.Millisecond)
		if err = sendSaveToFlashCommand(port); err == nil {
			time.Sleep(200 * time.Millisecond)
			if err = sendDeepSleepCommand(port); err == nil {
				time.Sleep(1000 * time.Millisecond)
				if err = sendWakeUpModuleCommand(port); err == nil {
					return
				}
			}
		}
	}

	log.Printf("failed to write command: %s\n", err)
}

func LightControlHandler(port *serial.Port) mqtt.MessageHandler {
	return func(client mqtt.Client, message mqtt.Message) {

		state := string(message.Payload())
		if state == "true" {
			lightOn(port)
		} else if state == "false" {
			lightOff(port)
		} else {
			log.Printf("invalid state received: %s", state)
		}
	}
}
