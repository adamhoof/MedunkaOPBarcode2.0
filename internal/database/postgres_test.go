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
		"POSTGRES_DB",
		"DB_TABLE_NAME",
		"POSTGRES_SSLMODE",
		"DB_USER_FILE",
		"DB_PASSWORD_FILE",
		"TLS_CA_PATH",
	}

	for _, key := range required {
		t.Setenv(key, "value")
	}

	missingKey := "POSTGRES_HOSTNAME"
	t.Setenv(missingKey, "")

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic when required env vars are missing")
		}
		message := r.(string)
		if !strings.Contains(message, missingKey) {
			t.Fatalf("expected panic to mention %s, got %s", missingKey, message)
		}
	}()

	_, _ = database.NewPostgres()
}
