package body

import (
	"encoding/json"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/helpers/xss"
)

func GetBody[T any](r *http.Request, u *T) error {
	if err := json.NewDecoder(r.Body).Decode(u); err != nil {
		return err
	}

	xss.SanitizeStruct(u)
	return nil
}
