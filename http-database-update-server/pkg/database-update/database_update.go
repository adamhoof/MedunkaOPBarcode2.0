package database_update

import (
	"encoding/json"
	"fmt"
	"github.com/adamhoof/MedunkaOPBarcode2.0/config"
	"github.com/adamhoof/MedunkaOPBarcode2.0/database"
	"io"
	"log"
	"net/http"
	"os"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func (e ErrorResponse) Error() string {
	return e.Message
}

type SuccessResponse struct {
	Message string `json:"message"`
}

func extractFileFromRequest(request *http.Request, conf *config.Config) (string, error) {
	receivedFile, _, err := request.FormFile("file")
	if err != nil {
		return "", fmt.Errorf("failed to read multipart file from request: %s", err)
	}
	defer func() {
		err = receivedFile.Close()
		if err != nil {
			log.Printf("failed to close received multipart file: %s\n", err)
		}
	}()

	tmpFile, err := os.CreateTemp(conf.HTTPDatabaseUpdate.TempFileLocation, "*.mdb")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %s", err)
	}

	if _, err = io.Copy(tmpFile, receivedFile); err != nil {
		return "", fmt.Errorf("failed to write temporary file: %s", err)
	}

	return tmpFile.Name(), nil
}

func handleError(w http.ResponseWriter, err error, statusCode int) {
	w.WriteHeader(statusCode)
	err = json.NewEncoder(w).Encode(ErrorResponse{Message: err.Error()})
	if err != nil {
		log.Printf("failed to encode message into json: %s\n", err)
	}
}
func HandleDatabaseUpdateRequest(conf *config.Config, handler database.DatabaseHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		tmpFileLocation, err := extractFileFromRequest(request, conf)
		if err != nil {
			handleError(w, err, http.StatusBadRequest)
			return
		}
		defer func() {
			err = os.Remove(tmpFileLocation)
			if err != nil {
				log.Printf("failed to remove temporary file: %s\n", err)
			}
		}()

		if err = handler.Connect(&conf.Database); err != nil {
			handleError(w, err, http.StatusInternalServerError)
			return
		}
		defer func() {
			err = handler.Disconnect()
			if err != nil {
				log.Printf("failed to disconnect from database: %s\n", err)
			}
		}()

		fields := []database.TableField{
			{Name: "name", Type: "TEXT"},
			{Name: "barcode", Type: "TEXT"},
			{Name: "price", Type: "TEXT"},
			{Name: "unit_of_measure", Type: "TEXT"},
			{Name: "unit_of_measure_koef", Type: "TEXT"},
			{Name: "stock", Type: "TEXT"},
		}

		if err = handler.DropTableIfExists(conf.HTTPDatabaseUpdate.TableName); err != nil {
			handleError(w, err, http.StatusInternalServerError)
			return
		}

		if err = handler.CreateTable(conf.HTTPDatabaseUpdate.TableName, fields); err != nil {
			handleError(w, err, http.StatusInternalServerError)
			return
		}

		if err = handler.ImportCSV(conf.HTTPDatabaseUpdate.TableName, conf.HTTPDatabaseUpdate.OutputCSVLocation, conf.HTTPDatabaseUpdate.Delimiter); err != nil {
			handleError(w, err, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		response := SuccessResponse{Message: "All done!"}

		if err = json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("failed to write to client: %s\n", err)
		}
	}
}
