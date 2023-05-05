package barcode

import (
	"bufio"
	"github.com/tarm/serial"
)

type ReaderHandler struct {
	reader *bufio.Reader
}

func (handler *ReaderHandler) ReadUntilDelimiter(delimiter byte) []byte {
	barcode, err := handler.reader.ReadBytes(delimiter)
	if err != nil {
		panic(err)
	}
	return barcode
}

func (handler *ReaderHandler) GetPort(port *serial.Port) {
	handler.reader = bufio.NewReader(port)
}
