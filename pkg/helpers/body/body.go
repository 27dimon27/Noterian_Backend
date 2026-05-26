package body

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/helpers/xss"
	"github.com/mailru/easyjson"
)

func GetBody(r *http.Request, data easyjson.Unmarshaler) error {
	if err := easyjson.UnmarshalFromReader(r.Body, data); err != nil {
		return err
	}

	xss.SanitizeStruct(data)
	return nil
}
