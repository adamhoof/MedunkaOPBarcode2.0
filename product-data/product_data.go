package product_data

type ProductData struct {
	Name              string `json:"name"`
	Price             string `json:"price"`
	Stock             string `json:"stock"`
	UnitOfMeasure     string `json:"unitOfMeasure"`
	UnitOfMeasureCoef string `json:"unitOfMeasureCoef"`
}

type ProductDataRequest struct {
	ClientTopic       string `json:"clientTopic"`
	Barcode           string `json:"barcode"`
	IncludeDiacritics bool   `json:"includeDiacritics"`
}
