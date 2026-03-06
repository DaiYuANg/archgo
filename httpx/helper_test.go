package httpx

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteJSON_MarshalError(t *testing.T) {
	w := httptest.NewRecorder()

	// channel documents related behavior.
	writeJSON(w, http.StatusAccepted, map[string]interface{}{"invalid": make(chan int)})

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, strings.ToLower(w.Body.String()), "internal server error")
}

func TestWriteJSON_Success(t *testing.T) {
	w := httptest.NewRecorder()

	writeJSON(w, http.StatusCreated, map[string]string{"message": "ok"})

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "\"message\":\"ok\"")
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestTypeNameOfNil(t *testing.T) { assert.Equal(t, "<nil>", typeNameOf(nil)) }
