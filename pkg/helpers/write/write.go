package write

import (
	"net/http"

	"github.com/mailru/easyjson"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func JSONResponse(w http.ResponseWriter, status int, data easyjson.Marshaler) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if data == nil {
		return
	}

	if _, err := easyjson.MarshalToWriter(data, w); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func JSONErrorResponse(w http.ResponseWriter, status int, err error) {
	JSONResponse(w, status, ErrorResponse{
		Error: err.Error(),
	})
}
