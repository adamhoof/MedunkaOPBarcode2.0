package product_data

type ProductData struct {
	Name              string
	Price             string
	Stock             string
	UnitOfMeasure     string
	UnitOfMeasureKoef string
	FirmwareUpdate    bool
}

type ProductDataRequest struct {
	ClientTopic       string `json:"clientTopic"`
	Barcode           string `json:"barcode"`
	IncludeDiacritics bool   `json:"includeDiacritics"`
}
