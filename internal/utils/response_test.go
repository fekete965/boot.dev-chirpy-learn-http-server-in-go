package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRespondWithNoContent(t *testing.T) {
	recorder := httptest.NewRecorder()

	RespondWithNoContent(recorder)

	if recorder.Code != http.StatusNoContent {
		t.Errorf("Expected %d status code, but received %d", http.StatusNoContent, recorder.Code)
	}

	currentContentType := recorder.Header().Get("Content-Type")
	expectedContentType := "text/plain; charset=utf-8"

	if currentContentType != expectedContentType {
		t.Errorf("Expected %s content type, but received %s", expectedContentType, currentContentType)
	}

	if recorder.Body.Len() != 0 {
		t.Errorf("Expected body to be empty, but received %s", recorder.Body.String())
	}
}

func TestRespondWithPlainText(t *testing.T) {
	recorder := httptest.NewRecorder()

	RespondWithPlainText(recorder, http.StatusOK, "Hello, there!")

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected %d status code, but received %d", http.StatusOK, recorder.Code)
	}

	currentContentType := recorder.Header().Get("Content-Type")
	expectedContentType := "text/plain; charset=utf-8"

	if currentContentType != expectedContentType {
		t.Errorf("Expected %s content type, but received %s", expectedContentType, currentContentType)
	}

	if recorder.Body.Len() == 0 {
		t.Errorf("Expected body not to be empty")
	}

	if recorder.Body.String() != "Hello, there!" {
		t.Errorf("Expected body to be %s, but received %s", "Hello, there!", recorder.Body.String())
	}
}

func TestRespondWithJSON(t *testing.T) {
	recorder := httptest.NewRecorder()

	type testPayload struct {
		Message string `json:"message"`
	}
	payload := testPayload{Message: "Hello, there!"}

	RespondWithJSON(recorder, http.StatusOK, payload)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected %d status code, but received %d", http.StatusOK, recorder.Code)
	}

	currentContentType := recorder.Header().Get("Content-Type")
	expectedContentType := "text/json; charset=utf-8"

	if currentContentType != expectedContentType {
		t.Errorf("Expected %s content type, but received %s", expectedContentType, currentContentType)
	}

	if recorder.Body.Len() == 0 {
		t.Errorf("Expected body not to be empty")
	}

	expectedBody := `{"message":"Hello, there!"}`	
	if recorder.Body.String() != expectedBody {
		t.Errorf("Expected body to be %s, but received %s", expectedBody, recorder.Body.String())
	}
}

func TestRespondWithJSONWithMarshallingError(t *testing.T) {
	recorder := httptest.NewRecorder()

	var payload chan string
	RespondWithJSON(recorder, http.StatusOK, payload)

	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("Expected %d status code, but received %d", http.StatusInternalServerError, recorder.Code)
	}

	currentContentType := recorder.Header().Get("Content-Type")
	expectedContentType := "text/plain; charset=utf-8"

	if currentContentType != expectedContentType {
		t.Errorf("Expected %s content type, but received %s", expectedContentType, currentContentType)
	}

	if recorder.Body.Len() == 0 {
		t.Errorf("Expected body not to be empty")
	}
}
