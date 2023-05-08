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
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
}

func (e ErrorResponse) Error() string {
	return e.Message
}

func extractFileFromRequest(request *http.Request, conf *config.Config) (string, error) {
	receivedFile, _, err := request.FormFile("file")
	if err != nil {
		return "", ErrorResponse{
			Message:    fmt.Sprintf("failed to read multipart file from request: %s", err),
			StatusCode: http.StatusBadRequest,
		}
	}
	defer receivedFile.Close()

	tmpFile, err := os.CreateTemp(conf.HTTPDatabaseUpdate.TempFileLocation, "*.mdb")
	if err != nil {
		return "", ErrorResponse{
			Message:    fmt.Sprintf("failed to create temporary file: %s", err),
			StatusCode: http.StatusInternalServerError,
		}
	}
	defer tmpFile.Close()

	if _, err = io.Copy(tmpFile, receivedFile); err != nil {
		return "", ErrorResponse{
			Message:    fmt.Sprintf("failed to write temporary file: %s", err),
			StatusCode: http.StatusInternalServerError,
		}
	}

	return tmpFile.Name(), nil
}

func handleError(w http.ResponseWriter, err error) {
	if err, ok := err.(ErrorResponse); ok {
		w.WriteHeader(err.StatusCode)
		json.NewEncoder(w).Encode(err)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Message:    fmt.Sprintf("An unexpected error occurred: %s", err),
			StatusCode: http.StatusInternalServerError,
		})
	}
}

func HandleDatabaseUpdateRequest(conf *config.Config, handler database.DatabaseHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		tmpFileName, err := extractFileFromRequest(r, conf)
		if err != nil {
			handleError(w, err)
			return
		}
		defer os.Remove(tmpFileName)

		if err = handler.Connect(&conf.Database); err != nil {
			handleError(w, ErrorResponse{
				Message:    fmt.Sprintf("failed to connect to database: %s", err),
				StatusCode: http.StatusInternalServerError,
			})
			return
		}
		defer handler.Disconnect()

		fields := []database.TableField{
			{Name: "name", Type: "TEXT"},
			{Name: "barcode", Type: "TEXT"},
			{Name: "price", Type: "TEXT"},
			{Name: "unit_of_measure", Type: "TEXT"},
			{Name: "unit_of_measure_koef", Type: "TEXT"},
			{Name: "stock", Type: "TEXT"},
		}

		if err = handler.DropTableIfExists(conf.HTTPDatabaseUpdate.TableName); err != nil {
			handleError(w, ErrorResponse{
				Message:    fmt.Sprintf("failed to drop database table: %s", err),
				StatusCode: http.StatusInternalServerError,
			})
			return
		}

		if err = handler.CreateTable(conf.HTTPDatabaseUpdate.TableName, fields); err != nil {
			handleError(w, ErrorResponse{
				Message:    fmt.Sprintf("failed to create database table: %s", err),
				StatusCode: http.StatusInternalServerError,
			})
			return
		}

		if err = handler.ImportCSV(conf.HTTPDatabaseUpdate.TableName, conf.HTTPDatabaseUpdate.OutputCSVLocation, conf.HTTPDatabaseUpdate.Delimiter); err != nil {
			handleError(w, ErrorResponse{
				Message:    fmt.Sprintf("failed to update database with csv file: %s", err),
				StatusCode: http.StatusInternalServerError,
			})
			return
		}

		response := struct {
			Message string `json:"message"`
		}{
			Message: "All done!",
		}

		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(response); err != nil {
			log.Println("failed to write to client")
		}
	}
}
