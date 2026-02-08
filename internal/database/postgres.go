package database

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/adamhoof/MedunkaOPBarcode2.0/internal/domain"
	"github.com/adamhoof/MedunkaOPBarcode2.0/internal/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
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

	cert, err := tls.LoadX509KeyPair(config.ClientCertPath, config.ClientKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load client key pair: %w", err)
	}

	caCert, err := os.ReadFile(config.CAPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA cert: %w", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append CA cert")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		ServerName:   config.Host,
		MinVersion:   tls.VersionTLS12,
	}

	connString := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s sslmode=require",
		config.Host,
		config.Port,
		config.User,
		config.Database,
	)

	pgxConfig, err := pgx.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	pgxConfig.TLSConfig = tlsConfig

	connStr := stdlib.RegisterConnConfig(pgxConfig)

	db, err := sql.Open("pgx", connStr)
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
	User            string
	Database        string
	TableName       string
	ClientCertPath  string
	ClientKeyPath   string
	CAPath          string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func loadPostgresConfig() postgresConfig {
	return postgresConfig{
		Host:            utils.GetEnvOrPanic("POSTGRES_HOST"),
		Port:            utils.GetEnvOrPanic("POSTGRES_PORT"),
		User:            utils.ReadSecretOrFail("POSTGRES_USER_FILE"),
		Database:        utils.GetEnvOrPanic("POSTGRES_DB"),
		TableName:       utils.GetEnvOrPanic("DB_TABLE_NAME"),
		ClientCertPath:  utils.GetEnvOrPanic("TLS_CLIENT_CERT_PATH"),
		ClientKeyPath:   utils.GetEnvOrPanic("TLS_CLIENT_KEY_PATH"),
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
	conn, err := p.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Close()

	return conn.Raw(func(driverConn any) error {
		stdlibConn, ok := driverConn.(*stdlib.Conn)
		if !ok {
			return errors.New("driver connection is not a *stdlib.Conn")
		}

		pgxConn := stdlibConn.Conn()

		source := &streamSource{stream: stream}
		_, err := pgxConn.CopyFrom(
			ctx,
			pgx.Identifier{p.tableName},
			[]string{"name", "barcode", "price", "unit_of_measure", "unit_of_measure_koef", "stock"},
			source,
		)
		if err != nil {
			return fmt.Errorf("copy failed: %w", err)
		}

		return nil
	})
}

type streamSource struct {
	stream  <-chan domain.Product
	current domain.Product
	err     error
}

func (s *streamSource) Next() bool {
	var ok bool
	s.current, ok = <-s.stream
	return ok
}

func (s *streamSource) Values() ([]any, error) {
	return []any{
		s.current.Name,
		s.current.Barcode,
		s.current.Price,
		s.current.UnitOfMeasure,
		s.current.UnitOfMeasureCoef,
		s.current.Stock,
	}, nil
}

func (s *streamSource) Err() error {
	return s.err
}
