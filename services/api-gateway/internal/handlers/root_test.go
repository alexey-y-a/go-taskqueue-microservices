package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRootHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()

	handler := RootHandler()
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	require.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var response RootResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err, "должен быть валидный JSON")

	require.Equal(t, "api-gateway", response.Service)
	require.Equal(t, "running", response.Status)
	require.Equal(t, "v1", response.Version)

}
