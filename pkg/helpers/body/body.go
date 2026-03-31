package body

import (
	"encoding/json"
	"net/http"
)

func GetBody[T any](r *http.Request, u *T) error {
	if err := json.NewDecoder(r.Body).Decode(u); err != nil {
		return err
	}
	return nil
}
