package file_parser

import (
	"bytes"
	"fmt"
	"github.com/adamhoof/MedunkaOPBarcode2.0/config"
	"log"
	"os/exec"
)

type MDBFileParser struct {
}

func (mdbFileParser *MDBFileParser) ToCSV(inputMDBLocation string, outputCSVLocation string) (err error) {
	conf, err := config.LoadConfig("/home/adamhoof/MedunkaOPBarcode2.0/Config.json")
	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command(conf.HTTPDatabaseUpdate.ShellMDBParserLocation, inputMDBLocation, outputCSVLocation)
	var stderr bytes.Buffer

	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run mdb_file_parser.sh: " + stderr.String())
	}

	return nil
}
