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

	file, err := os.Create(os.Getenv("HTTP_SERVER_CSV_FILE_PATH"))
	if err != nil {
		return "", fmt.Errorf("failed to create file: %s", err)
	}

	err = file.Chmod(0666)
	if err != nil {
		return "", fmt.Errorf("failed to change permissions of file %s: %s", file.Name(), err)
	}

	if _, err = io.Copy(file, receivedFile); err != nil {
		return "", fmt.Errorf("failed to write file %s: %s", file.Name(), err)
	}

	return file.Name(), nil
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

		fileLocation, err := extractFileFromRequest(request)
		if err != nil {
			handleError(w, err, http.StatusBadRequest)
			return
		}
		defer func() {
			err = os.Remove(fileLocation)
			if err != nil {
				log.Printf("failed to remove file: %s\n", err)
			}
		}()

		connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			os.Getenv("POSTGRES_HOSTNAME"),
			os.Getenv("POSTGRES_PORT"),
			os.Getenv("POSTGRES_USER"),
			os.Getenv("POSTGRES_PASSWORD"),
			os.Getenv("POSTGRES_DB"))

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

		if err = handler.ImportCSV(os.Getenv("DB_TABLE_NAME"), fileLocation, os.Getenv("DB_DELIMITER")); err != nil {
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
