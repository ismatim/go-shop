package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type errorWriter struct {
	http.ResponseWriter
}

func (e *errorWriter) Write(b []byte) (int, error) {
	return 0, io.ErrClosedPipe
}

func TestOKResponse(t *testing.T) {
	type sampleResponse struct {
		Message string `json:"message"`
	}

	sample := sampleResponse{Message: "Success"}

	t.Run("succesful http200 json response", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		OKResponse(recorder, sample)

		assert.Equal(t, http.StatusOK, recorder.Code, "Expected status code 200 OK")
		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"), "Expected Content-Type to be application/json")

		expected := `{"message":"Success"}`
		assert.JSONEq(t, expected, recorder.Body.String(), "Response body does not match expected")
	})
}

func TestErrorResponse(t *testing.T) {
	t.Run("json response for a given http status code", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		ErrorResponse(recorder, http.StatusInternalServerError, "Some error occurred")

		assert.Equal(t, http.StatusInternalServerError, recorder.Code, "Expected status code 500 Internal Server Error")
		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"), "Expected Content-Type to be application/json")

		expected := `{"error":"Some error occurred"}`
		assert.JSONEq(t, expected, recorder.Body.String(), "Response body does not match expected")
	})
}

func TestErrorResponse_EncodeError(t *testing.T) {
	rr := httptest.NewRecorder()
	ew := &errorWriter{rr}
	ErrorResponse(ew, http.StatusBadRequest, string([]byte{0xff, 0xfe, 0xfd})) // invalid UTF-8

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}
}

func TestOkResponse_EncodeError(t *testing.T) {
	rr := httptest.NewRecorder()
	ew := &errorWriter{rr}
	OKResponse(ew, string([]byte{0xff, 0xfe, 0xfd})) // invalid UTF-8

	resp := rr.Result()
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}
}
