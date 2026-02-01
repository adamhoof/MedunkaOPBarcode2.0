package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/adamhoof/MedunkaOPBarcode2.0/internal/domain"
	"github.com/adamhoof/MedunkaOPBarcode2.0/internal/utils"
	"github.com/lib/pq"
)

const (
	defaultQueryTimeout = 5 * time.Second
)

type Postgres struct {
	db        *sql.DB
	tableName string
	config    postgresConfig
}

func NewPostgres() (Handler, error) {
	config := loadPostgresConfig()

	connectionString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s sslrootcert=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.Database,
		config.SSLMode,
		config.CAPath,
	)

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("could not open connection: %w", err)
	}

	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)

	if err = db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	return &Postgres{db: db, tableName: config.TableName, config: config}, nil
}

type postgresConfig struct {
	Host            string
	Port            string
	Database        string
	TableName       string
	User            string
	Password        string
	SSLMode         string
	CAPath          string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func loadPostgresConfig() postgresConfig {
	return postgresConfig{
		Host:            utils.GetEnvOrPanic("POSTGRES_HOST"),
		Port:            utils.GetEnvOrPanic("POSTGRES_PORT"),
		Database:        utils.GetEnvOrPanic("POSTGRES_DB"),
		TableName:       utils.GetEnvOrPanic("DB_TABLE_NAME"),
		User:            utils.ReadSecretOrFail("POSTGRES_USER_FILE"),
		Password:        utils.ReadSecretOrFail("POSTGRES_PASSWORD_FILE"),
		SSLMode:         utils.GetEnvOrPanic("POSTGRES_SSLMODE"),
		CAPath:          utils.GetEnvOrPanic("TLS_CA_PATH"),
		MaxOpenConns:    utils.GetEnvAsInt("DB_MAX_OPEN_CONNS"),
		MaxIdleConns:    utils.GetEnvAsInt("DB_MAX_IDLE_CONNS"),
		ConnMaxLifetime: utils.GetEnvAsDuration("DB_CONN_MAX_LIFETIME"),
	}
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

func (p *Postgres) Close() error {
	if p.db == nil {
		return nil
	}
	if err := p.db.Close(); err != nil {
		return fmt.Errorf("failed to disconnect from db: %w", err)
	}
	return nil
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
	txn, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	stmt, err := txn.PrepareContext(ctx, pq.CopyIn(
		p.tableName,
		"name",
		"barcode",
		"price",
		"unit_of_measure",
		"unit_of_measure_koef",
		"stock",
	))
	if err != nil {
		_ = txn.Rollback()
		return fmt.Errorf("failed to prepare copy statement: %w", err)
	}
	defer stmt.Close()

	for {
		select {
		case <-ctx.Done():
			_ = txn.Rollback()
			return ctx.Err()
		case record, ok := <-stream:
			if !ok {
				if _, err := stmt.Exec(); err != nil {
					_ = txn.Rollback()
					return fmt.Errorf("failed to flush copy data: %w", err)
				}

				if err := txn.Commit(); err != nil {
					return fmt.Errorf("failed to commit transaction: %w", err)
				}
				return nil
			}

			if _, err := stmt.Exec(record.Name, record.Barcode, record.Price, record.UnitOfMeasure, record.UnitOfMeasureCoef, record.Stock); err != nil {
				_ = txn.Rollback()
				return fmt.Errorf("failed to buffer copy data: %w", err)
			}
		}
	}
}
