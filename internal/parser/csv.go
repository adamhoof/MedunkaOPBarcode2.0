package parser

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/adamhoof/MedunkaOPBarcode2.0/internal/domain"
)

type CSV struct{}

func NewCSV() (CatalogParser, error) {
	return &CSV{}, nil
}

func (parser *CSV) ParseStream(r io.Reader) (<-chan domain.Product, <-chan error) {
	dataCh := make(chan domain.Product)
	errCh := make(chan error, 1)

	go func() {
		defer close(dataCh)
		defer close(errCh)

		reader := csv.NewReader(r)
		reader.Comma = ';'
		reader.LazyQuotes = true
		reader.FieldsPerRecord = 6

		_, err := reader.Read()
		if err != nil {
			sendError(errCh, fmt.Errorf("failed to read csv header: %w", err))
			return
		}

		for {
			record, err := reader.Read()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				sendError(errCh, fmt.Errorf("failed to read csv row: %w", err))
				return
			}

			product := domain.Product{
				Name:              clean(record[0]),
				Barcode:           clean(record[1]),
				Price:             clean(record[2]),
				UnitOfMeasure:     clean(record[3]),
				UnitOfMeasureCoef: clean(record[4]),
				Stock:             clean(record[5]),
			}

			dataCh <- product
		}
	}()

	return dataCh, errCh
}

func clean(value string) string {
	val := strings.TrimSpace(value)
	return strings.Trim(val, "\"")
}

func sendError(errCh chan<- error, err error) {
	select {
	case errCh <- err:
	default:
	}
}
