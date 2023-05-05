package serial_communication

import "github.com/tarm/serial"

func OpenPort(config *serial.Config) (*serial.Port, error) {
	port, err := serial.OpenPort(config)
	return port, err
}
