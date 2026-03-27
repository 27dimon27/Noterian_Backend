package write

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONResponse(t *testing.T) {
	tests := []struct {
		name          string
		status        int
		data          interface{}
		expectedBody  interface{}
		checkResponse func(t *testing.T, body []byte)
	}{
		{
			name:   "success response with struct",
			status: http.StatusOK,
			data: struct {
				ID    string `json:"id"`
				Login string `json:"login"`
			}{
				ID:    "123",
				Login: "test123",
			},
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Equal(t, "123", response["id"])
				assert.Equal(t, "test123", response["login"])
			},
		},
		{
			name:   "success response with map",
			status: http.StatusCreated,
			data: map[string]string{
				"message": "created successfully",
			},
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]string
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Equal(t, "created successfully", response["message"])
			},
		},
		{
			name:   "success response with nil data",
			status: http.StatusNoContent,
			data:   nil,
			checkResponse: func(t *testing.T, body []byte) {
				assert.Equal(t, "", string(body))
			},
		},
		{
			name:   "error response with error",
			status: http.StatusBadRequest,
			data:   map[string]string{"error": "invalid input"},
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]string
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], "invalid input")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			JSONResponse(w, tt.status, tt.data)

			assert.Equal(t, tt.status, w.Code)
			assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}

func TestJSONErrorResponse(t *testing.T) {
	tests := []struct {
		name          string
		status        int
		err           error
		expectedError string
	}{
		{
			name:          "bad request error",
			status:        http.StatusBadRequest,
			err:           errors.New("invalid input"),
			expectedError: "invalid input",
		},
		{
			name:          "unauthorized error",
			status:        http.StatusUnauthorized,
			err:           errors.New("unauthorized"),
			expectedError: "unauthorized",
		},
		{
			name:          "not found error",
			status:        http.StatusNotFound,
			err:           errors.New("resource not found"),
			expectedError: "resource not found",
		},
		{
			name:          "internal server error",
			status:        http.StatusInternalServerError,
			err:           errors.New("internal error"),
			expectedError: "internal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			JSONErrorResponse(w, tt.status, tt.err)

			assert.Equal(t, tt.status, w.Code)
			assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

			var response map[string]string
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Contains(t, response["error"], tt.expectedError)
			assert.Len(t, response, 1)
		})
	}
}
