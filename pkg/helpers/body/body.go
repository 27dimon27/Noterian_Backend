package body

import (
	"encoding/json"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/dto"
)

func GetBody[T dto.SignInUser | dto.SignUpUser](r *http.Request, u *T) error {
	if err := json.NewDecoder(r.Body).Decode(u); err != nil {
		return err
	}
	return nil
}
