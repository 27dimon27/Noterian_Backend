package body

import (
	"encoding/json"
	"net/http"

	authDTO "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/dto"
	notesDTO "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/dto"
	profilesDTO "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/dto"
)

func GetBody[T authDTO.SignInUser | authDTO.SignUpUser | profilesDTO.Profile | notesDTO.NoteRequest](r *http.Request, u *T) error {
	if err := json.NewDecoder(r.Body).Decode(u); err != nil {
		return err
	}
	return nil
}
