package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/adamhoof/MedunkaOPBarcode2.0/internal/domain"
	_ "github.com/lib/pq"
)

const (
	defaultQueryTimeout = 5 * time.Second
)

type Postgres struct {
	db        *sql.DB
	tableName string
}

func NewPostgres() (Handler, error) {
	requiredEnv := []string{
		"POSTGRES_HOSTNAME",
		"POSTGRES_PORT",
		"POSTGRES_USER",
		"POSTGRES_PASSWORD",
		"POSTGRES_DB",
		"DB_TABLE_NAME",
	}

	missing := missingEnv(requiredEnv)
	if len(missing) > 0 {
		log.Printf("missing required env vars: %s", strings.Join(missing, ", "))
		return nil, fmt.Errorf("missing required env vars: %s", strings.Join(missing, ", "))
	}

	connectionString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("POSTGRES_HOSTNAME"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
	)

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("could not open connection: %w", err)
	}

	if err = db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	return &Postgres{db: db, tableName: os.Getenv("DB_TABLE_NAME")}, nil
}

func (p *Postgres) Close() error {
	if p.db == nil {
		return nil
	}
	if err := p.db.Close(); err != nil {
		return fmt.Errorf("failed to disconnect from db: %w", err)
	}
	return nil
}

func (p *Postgres) Fetch(ctx context.Context, barcode string) (*domain.Product, error) {
	if barcode == "" {
		return nil, errors.New("barcode cannot be empty")
	}
	if p.db == nil {
		return nil, errors.New("database connection is not initialized")
	}

	queryCtx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	query := fmt.Sprintf(
		"SELECT name, price, stock, unit_of_measure, unit_of_measure_koef FROM %s WHERE barcode = $1",
		p.tableName,
	)

	row := p.db.QueryRowContext(queryCtx, query, barcode)
	productData := domain.Product{}
	if err := row.Scan(&productData.Name, &productData.Price, &productData.Stock, &productData.UnitOfMeasure, &productData.UnitOfMeasureCoef); err != nil {
		return nil, fmt.Errorf("unable to fetch product data: %w", err)
	}

	return &productData, nil
}

func (p *Postgres) ImportCatalog(ctx context.Context, stream <-chan domain.Product) error {
	if p.db == nil {
		return errors.New("database connection is not initialized")
	}

	if err := p.dropTable(ctx); err != nil {
		return err
	}

	if err := p.createTable(ctx); err != nil {
		return err
	}

	return p.copyStream(ctx, stream)
}

func (p *Postgres) dropTable(ctx context.Context) error {
	queryCtx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	_, err := p.db.ExecContext(queryCtx, fmt.Sprintf("DROP TABLE IF EXISTS %s;", p.tableName))
	if err != nil {
		return fmt.Errorf("failed to drop table %s: %w", p.tableName, err)
	}
	return nil
}

func (p *Postgres) createTable(ctx context.Context) error {
	queryCtx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	query := fmt.Sprintf(`CREATE TABLE %s (
name TEXT,
barcode TEXT,
price TEXT,
unit_of_measure TEXT,
unit_of_measure_koef TEXT,
stock TEXT
);`, p.tableName)

	_, err := p.db.ExecContext(queryCtx, query)
	if err != nil {
		return fmt.Errorf("failed to create table %s: %w", p.tableName, err)
	}
	return nil
}

func (p *Postgres) copyStream(ctx context.Context, stream <-chan domain.Product) error {
	transaction, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	stmt, err := transaction.PrepareContext(ctx, fmt.Sprintf(
		"INSERT INTO %s (name, barcode, price, unit_of_measure, unit_of_measure_koef, stock) VALUES ($1, $2, $3, $4, $5, $6)",
		p.tableName,
	))
	if err != nil {
		_ = transaction.Rollback()
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer func() {
		_ = stmt.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			_ = transaction.Rollback()
			return ctx.Err()
		case record, ok := <-stream:
			if !ok {
				if err := transaction.Commit(); err != nil {
					return fmt.Errorf("failed to commit transaction: %w", err)
				}
				return nil
			}
			if _, err := stmt.ExecContext(ctx, record.Name, record.Barcode, record.Price, record.UnitOfMeasure, record.UnitOfMeasureCoef, record.Stock); err != nil {
				_ = transaction.Rollback()
				return fmt.Errorf("failed to insert product data: %w", err)
			}
		}
	}
}

func missingEnv(keys []string) []string {
	missing := make([]string, 0)
	for _, key := range keys {
		if strings.TrimSpace(os.Getenv(key)) == "" {
			missing = append(missing, key)
		}
	}
	return missing
}
