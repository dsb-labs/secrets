package api_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/davidsbond/passwords/internal/server/api"
)

func request(t *testing.T, method, url string, body any) *http.Request {
	t.Helper()

	buffer := bytes.NewBuffer(nil)
	require.NoError(t, json.NewEncoder(buffer).Encode(body))

	return httptest.NewRequest(method, url, buffer)
}

func decode[T any](t *testing.T, r io.Reader) T {
	t.Helper()

	var out T
	require.NoError(t, json.NewDecoder(r).Decode(&out))
	return out
}

func assertAPIError(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()

	err := decode[api.Error](t, w.Body)
	assert.EqualValues(t, w.Code, err.Code)
	assert.NotEmpty(t, err.Message)
}

func assertResponse[T any](t *testing.T, w *httptest.ResponseRecorder, expected T) {
	t.Helper()

	actual := decode[T](t, w.Body)
	assert.EqualValues(t, expected, actual)
}
