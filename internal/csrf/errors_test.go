package csrf

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrors(t *testing.T) {
	t.Run("error messages are in Russian", func(t *testing.T) {
		assert.Equal(t, "CSRF-токен отсутствует", ErrCSRFTokenMissing.Error())
		assert.Equal(t, "Невалидный CSRF-токен", ErrCSRFTokenInvalid.Error())
		assert.Equal(t, "Не удалось сгенерировать CSRF-токен", ErrFailedToGenerateCSRFToken.Error())
	})

	t.Run("errors are comparable", func(t *testing.T) {
		err1 := ErrCSRFTokenMissing
		err2 := ErrCSRFTokenMissing
		err3 := ErrCSRFTokenInvalid

		assert.True(t, errors.Is(err1, err2))
		assert.False(t, errors.Is(err1, err3))
	})

	t.Run("errors can be wrapped", func(t *testing.T) {
		wrappedErr := errors.New("some error")
		combinedErr := errors.Join(ErrCSRFTokenInvalid, wrappedErr)

		assert.True(t, errors.Is(combinedErr, ErrCSRFTokenInvalid))
		assert.True(t, errors.Is(combinedErr, wrappedErr))
	})
}
