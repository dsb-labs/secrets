package api_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
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
