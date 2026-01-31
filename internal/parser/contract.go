package parser

import (
	"io"

	"github.com/adamhoof/MedunkaOPBarcode2.0/internal/domain"
)

// CatalogParser defines a parser that can stream product data from a catalog source.
type CatalogParser interface {
	ParseStream(r io.Reader) (<-chan domain.Product, <-chan error)
}
