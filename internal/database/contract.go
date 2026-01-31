package database

import (
	"context"

	"github.com/adamhoof/MedunkaOPBarcode2.0/internal/domain"
)

// Handler defines the contract for any database implementation.
type Handler interface {
	Close() error
	Fetch(ctx context.Context, barcode string) (*domain.Product, error)
	ImportCatalog(ctx context.Context, stream <-chan domain.Product) error
}
