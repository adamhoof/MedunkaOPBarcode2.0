package parser

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/adamhoof/MedunkaOPBarcode2.0/internal/domain"
)

type MMDB struct {
	table string
}

func NewMMDB() (CatalogParser, error) {
	value := strings.TrimSpace(os.Getenv("MMDB_TABLE_NAME"))
	if value == "" {
		return nil, fmt.Errorf("MMDB_TABLE_NAME is not set")
	}
	return &MMDB{table: value}, nil
}

func (parser *MMDB) ParseStream(r io.Reader) (<-chan domain.Product, <-chan error) {
	dataCh := make(chan domain.Product)
	errCh := make(chan error, 1)

	go func() {
		defer close(dataCh)
		defer close(errCh)

		tempFile, err := os.CreateTemp("", "catalog-*.mdb")
		if err != nil {
			sendError(errCh, fmt.Errorf("failed to create temp mdb file: %w", err))
			return
		}
		defer func() {
			_ = os.Remove(tempFile.Name())
		}()

		if _, err = io.Copy(tempFile, r); err != nil {
			sendError(errCh, fmt.Errorf("failed to write temp mdb file: %w", err))
			return
		}

		if err = tempFile.Close(); err != nil {
			sendError(errCh, fmt.Errorf("failed to close temp mdb file: %w", err))
			return
		}

		cmd := exec.Command("mdb-export", "--delimiter=;", "--quote=\"", "--escape-invisible", tempFile.Name(), parser.table)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			sendError(errCh, fmt.Errorf("failed to capture mdb-export output: %w", err))
			return
		}

		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		if err = cmd.Start(); err != nil {
			sendError(errCh, fmt.Errorf("failed to start mdb-export: %w", err))
			return
		}

		reader := csv.NewReader(stdout)
		reader.Comma = ';'
		reader.LazyQuotes = true
		reader.FieldsPerRecord = -1

		headers, err := reader.Read()
		if err != nil {
			_ = cmd.Process.Kill()
			sendError(errCh, fmt.Errorf("failed to read mdb header: %w", err))
			return
		}

		mapping, err := MapRequiredColumns(headers)
		if err != nil {
			_ = cmd.Process.Kill()
			sendError(errCh, err)
			return
		}

		for {
			record, err := reader.Read()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				_ = cmd.Process.Kill()
				sendError(errCh, fmt.Errorf("failed to read mdb row: %w", err))
				return
			}

			product, err := ParseRecord(record, mapping)
			if err != nil {
				_ = cmd.Process.Kill()
				sendError(errCh, err)
				return
			}

			dataCh <- product
		}

		if err = cmd.Wait(); err != nil {
			sendError(errCh, fmt.Errorf("mdb-export failed: %s", strings.TrimSpace(stderr.String())))
		}
	}()

	return dataCh, errCh
}

type ColumnMapping struct {
	nameIdx              int
	barcodeIdx           int
	priceIdx             int
	unitOfMeasureIdx     int
	unitOfMeasureCoefIdx int
	stockIdx             int
}

func MapRequiredColumns(headers []string) (ColumnMapping, error) {
	required := map[string]func(mapping *ColumnMapping, index int){
		"nazev":     func(m *ColumnMapping, index int) { m.nameIdx = index },
		"ean":       func(m *ColumnMapping, index int) { m.barcodeIdx = index },
		"prodejdph": func(m *ColumnMapping, index int) { m.priceIdx = index },
		"mj2":       func(m *ColumnMapping, index int) { m.unitOfMeasureIdx = index },
		"mj2koef":   func(m *ColumnMapping, index int) { m.unitOfMeasureCoefIdx = index },
		"stavz":     func(m *ColumnMapping, index int) { m.stockIdx = index },
	}

	mapping := ColumnMapping{
		nameIdx:              -1,
		barcodeIdx:           -1,
		priceIdx:             -1,
		unitOfMeasureIdx:     -1,
		unitOfMeasureCoefIdx: -1,
		stockIdx:             -1,
	}

	for idx, header := range headers {
		normalized := NormalizeHeader(header)
		if setter, ok := required[normalized]; ok {
			setter(&mapping, idx)
		}
	}

	missing := make([]string, 0)
	if mapping.nameIdx == -1 {
		missing = append(missing, "Nazev")
	}
	if mapping.barcodeIdx == -1 {
		missing = append(missing, "EAN")
	}
	if mapping.priceIdx == -1 {
		missing = append(missing, "ProdejDPH")
	}
	if mapping.unitOfMeasureIdx == -1 {
		missing = append(missing, "MJ2")
	}
	if mapping.unitOfMeasureCoefIdx == -1 {
		missing = append(missing, "MJ2Koef")
	}
	if mapping.stockIdx == -1 {
		missing = append(missing, "StavZ")
	}

	if len(missing) > 0 {
		return ColumnMapping{}, fmt.Errorf("missing required columns in MDB header: %s", strings.Join(missing, ", "))
	}

	return mapping, nil
}

func NormalizeHeader(header string) string {
	value := strings.TrimSpace(header)
	value = strings.Trim(value, "\"")
	value = strings.ToLower(value)
	value = strings.ReplaceAll(value, " ", "")
	return value
}

func ParseRecord(record []string, mapping ColumnMapping) (domain.Product, error) {
	requiredIndex := func(index int, name string) (string, error) {
		if index < 0 || index >= len(record) {
			return "", fmt.Errorf("missing column %s in record", name)
		}
		value := strings.TrimSpace(record[index])
		value = strings.Trim(value, "\"")
		return value, nil
	}

	name, err := requiredIndex(mapping.nameIdx, "Nazev")
	if err != nil {
		return domain.Product{}, err
	}
	barcode, err := requiredIndex(mapping.barcodeIdx, "EAN")
	if err != nil {
		return domain.Product{}, err
	}
	price, err := requiredIndex(mapping.priceIdx, "ProdejDPH")
	if err != nil {
		return domain.Product{}, err
	}
	unit, err := requiredIndex(mapping.unitOfMeasureIdx, "MJ2")
	if err != nil {
		return domain.Product{}, err
	}
	coef, err := requiredIndex(mapping.unitOfMeasureCoefIdx, "MJ2Koef")
	if err != nil {
		return domain.Product{}, err
	}
	stock, err := requiredIndex(mapping.stockIdx, "StavZ")
	if err != nil {
		return domain.Product{}, err
	}

	return domain.Product{
		Name:              name,
		Barcode:           barcode,
		Price:             price,
		UnitOfMeasure:     unit,
		UnitOfMeasureCoef: coef,
		Stock:             stock,
	}, nil
}

func sendError(errCh chan<- error, err error) {
	select {
	case errCh <- err:
	default:
	}
}
