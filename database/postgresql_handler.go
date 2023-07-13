package database

import (
	"database/sql"
	"fmt"
	product_data "github.com/adamhoof/MedunkaOPBarcode2.0/product-data"
	_ "github.com/lib/pq"
	"log"
	"strings"
)

type PostgreSQLHandler struct {
	db *sql.DB
}

func (handler *PostgreSQLHandler) Connect(connectionString string) (err error) {
	handler.db, err = sql.Open("postgres", connectionString)
	if err != nil {
		return fmt.Errorf("could not open connection %s", err)
	}
	return handler.db.Ping()
}

func (handler *PostgreSQLHandler) Disconnect() (err error) {
	err = handler.db.Close()
	if err != nil {
		return fmt.Errorf("failed to disconnect from db: %s", err)
	}
	return nil
}

func (handler *PostgreSQLHandler) FetchProductData(tableName string, barcode string) (productData product_data.ProductData, err error) {
	row := handler.db.QueryRow(fmt.Sprintf("SELECT name, price, stock, unit_of_measure, unit_of_measure_koef FROM %s WHERE barcode = '%s';", tableName, barcode))
	err = row.Scan(&productData.Name, &productData.Price, &productData.Stock, &productData.UnitOfMeasure, &productData.UnitOfMeasureCoef)
	if err != nil {
		log.Printf("unable to unpack row into product data struct: %s", err)
	}
	return productData, err
}

func (handler *PostgreSQLHandler) DropTableIfExists(tableName string) (err error) {
	_, err = handler.db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s;", tableName))
	if err != nil {
		return fmt.Errorf("failed to execute db statement %s", err)
	}
	return err
}

func (handler *PostgreSQLHandler) CreateTable(tableName string, fields []TableField) (err error) {
	queryBuilder := strings.Builder{}
	queryBuilder.WriteString("CREATE TABLE ")
	queryBuilder.WriteString(tableName)
	queryBuilder.WriteString(" (")

	for i, field := range fields {
		queryBuilder.WriteString(field.Name)
		queryBuilder.WriteString(" ")
		queryBuilder.WriteString(field.Type)

		if field.Constraint != "" {
			queryBuilder.WriteString(" ")
			queryBuilder.WriteString(field.Constraint)
		}

		if i < len(fields)-1 {
			queryBuilder.WriteString(", ")
		}
	}
	queryBuilder.WriteString(")")

	_, err = handler.db.Exec(queryBuilder.String())
	if err != nil {
		return fmt.Errorf("error creating table %s: %s", tableName, err)
	}
	return nil
}

func (handler *PostgreSQLHandler) ImportCSV(tableName string, filePath string, delimiter string) (err error) {
	_, err = handler.db.Exec(fmt.Sprintf("COPY %s FROM '%s' WITH NULL AS E'\\'\\'' DELIMITER '%s' CSV HEADER;", tableName, filePath, delimiter))
	if err != nil {
		return fmt.Errorf("unable to import csv table: %s", err)
	}
	return nil
}
