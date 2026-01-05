package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandleHealthCheck(t *testing.T) {
	recorder := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "api/healthz", nil)
	require.NoError(t, err)

	HandleHealthCheck.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "OK", recorder.Body.String())
}
