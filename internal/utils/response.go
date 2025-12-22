package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func RespondWithNoContent(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNoContent)
}

func RespondWithPlainText(w http.ResponseWriter, statusCode int, text string) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(statusCode)
	w.Write([]byte(text))
}

func RespondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	marshalledData, err := json.Marshal(data)
	if err != nil {
		errorMessage := fmt.Sprintf("error marshalling response: %v", err)
		
		RespondWithPlainText(w, http.StatusInternalServerError, errorMessage)
		return
	}

	w.Header().Add("Content-Type", "text/json; charset=utf-8")
	w.WriteHeader(statusCode)
	w.Write(marshalledData)
}
