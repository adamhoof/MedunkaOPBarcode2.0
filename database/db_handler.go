package database

import (
	product_data "github.com/adamhoof/MedunkaOPBarcode2.0/product-data"
)

type DatabaseHandler interface {
	Connect(connectionString string) (err error)
	Disconnect() (err error)
	FetchProductData(tableName string, barcode string) (productData product_data.ProductData, err error)
	DropTableIfExists(tableName string) (err error)
	CreateTable(tableName string, tableFields []TableField) (err error)
	ImportCSV(tableName string, filePath string, delimiter string) (err error)
}
