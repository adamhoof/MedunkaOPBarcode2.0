package parser

import (
	"fmt"
	"path/filepath"
	"strings"
)

// New creates the appropriate parser based on the file extension.
func New(filename string) (CatalogParser, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".mmdb", ".mdb":
		return NewMMDB()
	default:
		return nil, fmt.Errorf("no parser found for extension: %s", ext)
	}
}
