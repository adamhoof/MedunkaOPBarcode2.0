package response_handler

import (
	cli_artist "MedunkaOpBarcodeMQTT/OpBarcodeRPI/pkg/CLIArtist"
	product_data "MedunkaOpBarcodeMQTT/ProductData"
	"encoding/json"
	"fmt"
	typeconv "github.com/adamhoof/GolangTypeConvertorWrapper/pkg"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"strings"
)

func GetProductDataResponse() mqtt.MessageHandler {
	return func(client mqtt.Client, message mqtt.Message) {

		productData := product_data.Object{}
		err := json.Unmarshal(message.Payload(), &productData)
		if err != nil {
			fmt.Println("error when unpacking data into product data struct")
		}

		if productData.Name == "" {
			fmt.Println("empty name, check if product exists in database...")
			return
		}

		strPriceWithoutSuffix := strings.ReplaceAll(productData.Price, ".00 Kč", "")

		cli_artist.PrintSpaces(1)
		cli_artist.PrintStyledText(cli_artist.ItalicWhite, productData.Name)
		cli_artist.PrintSpaces(2)
		cli_artist.PrintStyledText(cli_artist.BoldRed, fmt.Sprintf("Cena za ks: %s Kč", strPriceWithoutSuffix))
		cli_artist.PrintSpaces(4)

		if productData.UnitOfMeasureKoef == "" {
			cli_artist.PrintStyledText(cli_artist.ItalicWhite, "Stock: "+productData.Stock)
			return
		}

		pricePerUnitOfMeasure := typeconv.StringToFloat(productData.Price) * typeconv.StringToFloat(productData.UnitOfMeasureKoef)
		strPricePerUnitOfMeasure := typeconv.FloatToString(pricePerUnitOfMeasure)

		cli_artist.PrintStyledText(cli_artist.ItalicWhite, fmt.Sprintf("Přepočet na %s: %s Kč", productData.UnitOfMeasure, strPricePerUnitOfMeasure))
		cli_artist.PrintSpaces(1)
		cli_artist.PrintStyledText(cli_artist.ItalicWhite, "Stock: "+productData.Stock)
	}
}
