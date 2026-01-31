package parser_test

import (
	"os"
	"testing"

	"github.com/adamhoof/MedunkaOPBarcode2.0/internal/parser"
)

func TestNewMMDBMissing(t *testing.T) {
	if err := os.Unsetenv("MMDB_TABLE_NAME"); err != nil {
		t.Fatalf("failed to unset env: %s", err)
	}

	if _, err := parser.NewMMDB(); err == nil {
		t.Fatal("expected error when MMDB_TABLE_NAME is not set")
	}
}

func TestMapRequiredColumnsSuccess(t *testing.T) {
	headers := []string{"Nazev", "EAN", "ProdejDPH", "MJ2", "MJ2Koef", "StavZ"}

	if _, err := parser.MapRequiredColumns(headers); err != nil {
		t.Fatalf("expected mapping to succeed, got error: %s", err)
	}
}

func TestMapRequiredColumnsMissing(t *testing.T) {
	if _, err := parser.MapRequiredColumns([]string{"Nazev", "EAN"}); err == nil {
		t.Fatal("expected error when required columns are missing")
	}
}

func TestParseRecordTrimsQuotes(t *testing.T) {
	headers := []string{"Nazev", "EAN", "ProdejDPH", "MJ2", "MJ2Koef", "StavZ"}
	mapping, err := parser.MapRequiredColumns(headers)
	if err != nil {
		t.Fatalf("unexpected mapping error: %s", err)
	}

	record := []string{"\"Name\"", "123", "\"10.50\"", "ks", "1", "\"5\""}
	product, err := parser.ParseRecord(record, mapping)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if product.Name != "Name" || product.Price != "10.50" || product.Stock != "5" {
		t.Fatalf("unexpected parsed values: %+v", product)
	}
}
