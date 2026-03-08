package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHealthHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/healthz", nil)
	require.NoError(t, err, "не должно быть ошибки при создании запроса")

	rr := httptest.NewRecorder()

	handler := HealthHandler()
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, "должен быть статус 200 OK")

	require.Equal(t, "ok", rr.Body.String(), "тело ответа должно быть 'ok'")

	require.Equal(t, "text/plain", rr.Header().Get("Content-Type"), "должен быть заголовок Content-Type: text/plain")

}

func TestRootHandler_MethodNotAllowed(t *testing.T) {
	req, err := http.NewRequest("POST", "/healthz", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()

	handler := HealthHandler()
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, "POST запрос должен работать как GET")
}
