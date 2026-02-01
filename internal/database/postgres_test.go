package database_test

import (
	"strings"
	"testing"

	"github.com/adamhoof/MedunkaOPBarcode2.0/internal/database"
)

func TestNewPostgresMissingEnv(t *testing.T) {
	required := []string{
		"POSTGRES_HOST",
		"POSTGRES_PORT",
		"POSTGRES_DB",
		"DB_TABLE_NAME",
		"POSTGRES_SSLMODE",
		"POSTGRES_USER_FILE",
		"POSTGRES_PASSWORD_FILE",
		"TLS_CA_PATH",
		"DB_MAX_OPEN_CONNS",
		"DB_MAX_IDLE_CONNS",
		"DB_CONN_MAX_LIFETIME",
	}

	for _, key := range required {
		t.Setenv(key, "value")
	}

	missingKey := "POSTGRES_HOST"
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
