package database

import "database/sql"

type DatabaseHandler interface {
	Connect(config *string) (err error)
	ExecuteStatement(statement string) (err error)
	QueryProductDataRow(query string, barcode string) (row *sql.Row)
}
