package mqtt_handlers

import (
	"encoding/json"
	"fmt"
	"github.com/adamhoof/MedunkaOPBarcode2.0/op-barcode-rpi/pkg/cli_artist"
	product_data "github.com/adamhoof/MedunkaOPBarcode2.0/product-data"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
	"regexp"
	"strconv"
)

func ProductDataResponseHandler() mqtt.MessageHandler {
	return func(client mqtt.Client, message mqtt.Message) {

		productData := product_data.ProductData{}
		err := json.Unmarshal(message.Payload(), &productData)
		if err != nil {
			log.Printf("error when unpacking data into product data struct: %s\n", err)
		}

		if productData.Name == "" {
			log.Println("empty name, check if product exists in database...")
			return
		}

		if productData.Price == "" {
			log.Println("empty price, check if product exists in database...")
			return
		}
		characterFilterRegex := regexp.MustCompile(`[^0-9.,]`)
		floatPrice, err := strconv.ParseFloat(characterFilterRegex.ReplaceAllString(productData.Price, ""), 64)
		if err != nil {
			log.Println(err)
			return
		}

		floatPricePerUnitOfMeasureCoef, unitOfMeasureCoefErr := strconv.ParseFloat(productData.UnitOfMeasureCoef, 64)

		cli_artist.PrintSpaces(1)
		cli_artist.PrintStyledText(cli_artist.ItalicWhite(), productData.Name)
		cli_artist.PrintSpaces(2)
		cli_artist.PrintStyledText(cli_artist.BoldRed(), fmt.Sprintf("Cena za ks: %.2f Kč", floatPrice))
		cli_artist.PrintSpaces(4)

		if productData.UnitOfMeasureCoef == "" || unitOfMeasureCoefErr != nil {
			cli_artist.PrintStyledText(cli_artist.ItalicWhite(), fmt.Sprintf("Stock: %s", productData.Stock))
			return
		}

		pricePerUnitOfMeasure := floatPrice * floatPricePerUnitOfMeasureCoef
		cli_artist.PrintStyledText(cli_artist.ItalicWhite(), fmt.Sprintf("Přepočet na %s:\n %.2f Kč", productData.UnitOfMeasure, pricePerUnitOfMeasure))
		cli_artist.PrintSpaces(3)
		cli_artist.PrintStyledText(cli_artist.ItalicWhite(), fmt.Sprintf("Stock: %s", productData.Stock))
		cli_artist.PrintSpaces(2)
	}
}
