package auth

import "testing"

func TestErrorsAreDistinctAndNonEmpty(t *testing.T) {
	errs := map[string]error{
		"ErrInvalidInput":     ErrInvalidInput,
		"ErrInternal":         ErrInternal,
		"ErrUnauthorized":     ErrUnauthorized,
		"ErrMethodNotAllowed": ErrMethodNotAllowed,
		"ErrBadCredentials":   ErrBadCredentials,
		"ErrInvalidUsername":  ErrInvalidUsername,
		"ErrInvalidPassword":  ErrInvalidPassword,
		"ErrUserExist":        ErrUserExist,
		"ErrUserNotExist":     ErrUserNotExist,
		"ErrTokenCreation":    ErrTokenCreation,
		"ErrInvalidUserID":    ErrInvalidUserID,
	}

	seen := make(map[string]string, len(errs))
	for name, err := range errs {
		if err == nil {
			t.Errorf("%s is nil", name)
			continue
		}
		msg := err.Error()
		if msg == "" {
			t.Errorf("%s has empty message", name)
		}
		if prev, ok := seen[msg]; ok {
			t.Errorf("duplicate error message %q for %s and %s", msg, prev, name)
		}
		seen[msg] = name
	}
}
