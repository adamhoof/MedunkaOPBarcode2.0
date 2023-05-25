package database_update

import (
	"encoding/json"
	"fmt"
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

func extractFileFromRequest(request *http.Request) (string, error) {
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

	tmpFile, err := os.CreateTemp(os.Getenv("CSV_OUTPUT_PATH"), "*.csv")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %s", err)
	}

	err = tmpFile.Chmod(0666)
	if err != nil {
		return "", fmt.Errorf("failed to change permissions of temporary file: %s", err)
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
func HandleDatabaseUpdateRequest(handler database.DatabaseHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		updateFileLocation, err := extractFileFromRequest(request)
		if err != nil {
			handleError(w, err, http.StatusBadRequest)
			return
		}
		defer func() {
			err = os.Remove(updateFileLocation)
			if err != nil {
				log.Printf("failed to remove temporary file: %s\n", err)
			}
		}()

		connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"))

		if err = handler.Connect(connectionString); err != nil {
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

		if err = handler.DropTableIfExists(os.Getenv("DB_TABLE_NAME")); err != nil {
			handleError(w, err, http.StatusInternalServerError)
			return
		}

		if err = handler.CreateTable(os.Getenv("DB_TABLE_NAME"), fields); err != nil {
			handleError(w, err, http.StatusInternalServerError)
			return
		}

		if err = handler.ImportCSV(os.Getenv("DB_TABLE_NAME"), updateFileLocation, os.Getenv("DB_DELIMITER")); err != nil {
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
