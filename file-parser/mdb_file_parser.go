package file_parser

import (
	"bytes"
	"fmt"
	"os/exec"
)

type MDBFileParser struct {
}

func (mdbFileParser *MDBFileParser) ToCSV(inputMDBLocation string, outputCSVLocation string, helperParserLocation string) (err error) {
	cmd := exec.Command(helperParserLocation, inputMDBLocation, outputCSVLocation)
	var stderr bytes.Buffer

	cmd.Stderr = &stderr

	if err = cmd.Run(); err != nil {
		return fmt.Errorf("failed to run mdb_file_parser.sh: " + stderr.String())
	}

	return nil
}
