package database

import (
	"fmt"
	"strings"
)

// New creates a database handler based on the requested driver type.
func New(dbType string) (Handler, error) {
	switch strings.ToLower(dbType) {
	case "postgres":
		return NewPostgres()
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}
