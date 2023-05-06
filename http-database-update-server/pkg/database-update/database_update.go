package database_update

import (
	"fmt"
	"github.com/adamhoof/MedunkaOPBarcode2.0/config"
	"github.com/adamhoof/MedunkaOPBarcode2.0/database"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type ErrorAndStatusCode struct {
	err        error
	statusCode int
}

func extractFileFromRequest(request *http.Request, conf *config.Config) (errorAndStatusCode ErrorAndStatusCode, cleanupFunc func()) {
	receivedFile, _, err := request.FormFile("file")
	if err != nil {
		return ErrorAndStatusCode{err: fmt.Errorf("failed to read multipart file from request: %s", err), statusCode: http.StatusBadRequest}, nil
	}

	tmpFile, err := os.CreateTemp(conf.HTTPDatabaseUpdate.TempFileLocation, "*.mdb")
	if err != nil {
		return ErrorAndStatusCode{err: fmt.Errorf("failed to create temporary file: %s", err), statusCode: http.StatusInternalServerError}, nil
	}

	if _, err = io.Copy(tmpFile, receivedFile); err != nil {
		if errClose := tmpFile.Close(); errClose != nil {
			log.Printf("failed to close temporary file: %v", errClose)
		}
		if errRemove := os.Remove(tmpFile.Name()); errRemove != nil {
			log.Printf("failed to remove temporary file: %v", errRemove)
		}
		return ErrorAndStatusCode{err: fmt.Errorf("failed to write temporary file: %s", err), statusCode: http.StatusInternalServerError}, nil
	}

	cleanupFunc = func() {
		if err = tmpFile.Close(); err != nil {
			log.Printf("failed to close temporary file: %v", err)
		}
		if err = os.Remove(tmpFile.Name()); err != nil {
			log.Printf("failed to remove temporary file: %v", err)
		}
		if err = receivedFile.Close(); err != nil {
			log.Printf("failed to close received multipart file: %v", err)
		}
	}

	return ErrorAndStatusCode{}, cleanupFunc
}

func HandleDatabaseUpdateRequest(conf *config.Config, handler database.DatabaseHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var responseBuilder strings.Builder

		errorAndStatusCode, cleanupFunc := extractFileFromRequest(r, conf)
		defer cleanupFunc()
		if errorAndStatusCode.err != nil {
			http.Error(w, errorAndStatusCode.err.Error(), errorAndStatusCode.statusCode)
			return
		}

		responseBuilder.WriteString("successfully extracted file from request\n")
		responseBuilder.WriteString("successfully downloaded file\n")

		if err := handler.Connect(&conf.Database); err != nil {
			http.Error(w, "failed to connect to database", http.StatusInternalServerError)
			return
		}

		fields := []database.TableField{
			{Name: "name", Type: "TEXT"},
			{Name: "barcode", Type: "TEXT"},
			{Name: "price", Type: "TEXT"},
			{Name: "unit_of_measure", Type: "TEXT"},
			{Name: "unit_of_measure_koef", Type: "TEXT"},
			{Name: "stock", Type: "TEXT"},
		}

		if err := handler.DropTableIfExists(conf.HTTPDatabaseUpdate.TableName); err != nil {
			http.Error(w, "failed to drop database table", http.StatusInternalServerError)
			return
		}
		responseBuilder.WriteString("successfully dropped database table\n")

		if err := handler.CreateTable(conf.HTTPDatabaseUpdate.TableName, fields); err != nil {
			http.Error(w, "failed to create database table", http.StatusInternalServerError)
			return
		}
		responseBuilder.WriteString("successfully created database table\n")

		if err := handler.ImportCSV(conf.HTTPDatabaseUpdate.TableName, conf.HTTPDatabaseUpdate.OutputCSVLocation, conf.HTTPDatabaseUpdate.Delimiter); err != nil {
			http.Error(w, "failed to update database with csv file", http.StatusInternalServerError)
			return
		}
		responseBuilder.WriteString("successfully imported csv file as a database table\n")

		if err := handler.Disconnect(); err != nil {
			http.Error(w, "failed to disconnect from database", http.StatusInternalServerError)
			return
		}
		responseBuilder.WriteString("successfully disconnected from database\n")
		responseBuilder.WriteString("All done!")

		w.WriteHeader(http.StatusOK)
		if _, err := fmt.Fprint(w, responseBuilder.String()); err != nil {
			log.Println("failed to write to client")
		}
	}
}
