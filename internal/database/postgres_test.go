package database_test

import (
	"strings"
	"testing"

	"github.com/adamhoof/MedunkaOPBarcode2.0/internal/database"
)

func TestNewPostgresMissingEnv(t *testing.T) {
	required := []string{
		"POSTGRES_HOSTNAME",
		"POSTGRES_PORT",
		"POSTGRES_USER",
		"POSTGRES_PASSWORD",
		"POSTGRES_DB",
		"DB_TABLE_NAME",
	}

	for _, key := range required {
		t.Setenv(key, "value")
	}

	missingKey := "POSTGRES_HOSTNAME"
	t.Setenv(missingKey, "")

	_, err := database.NewPostgres()
	if err == nil {
		t.Fatal("expected error when required env vars are missing")
	}
	if !strings.Contains(err.Error(), missingKey) {
		t.Fatalf("expected error to mention %s, got %s", missingKey, err)
	}
}
