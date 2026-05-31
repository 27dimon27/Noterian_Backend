package usecase

import (
	"testing"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/grpcclient/mocks"
	"go.uber.org/mock/gomock"
)

func TestUsernameValidation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	u := newUsecase(t, mocks.NewMockProfilesServiceClient(ctrl))

	tests := []struct {
		name     string
		username string
		valid    bool
	}{
		{"simple", "alice", true},
		{"digits", "alice123", true},
		{"mixed letters", "AliceБоб", true},
		{"valid separators", "al_ice.test", true},
		{"leading underscore", "_alice", false},
		{"leading dot", ".alice", false},
		{"trailing underscore", "alice_", false},
		{"trailing dot", "alice.", false},
		{"double underscore", "al__ice", false},
		{"double dot", "al..ice", false},
		{"underscore dot", "al_.ice", false},
		{"dot underscore", "al._ice", false},
		{"invalid chars", "alice!", false},
		{"space", "al ice", false},
		{"empty", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := u.validate.Var(tc.username, "required,username")
			if (err == nil) != tc.valid {
				t.Errorf("for %q expected valid=%v, got err=%v", tc.username, tc.valid, err)
			}
		})
	}
}

func TestPasswordValidation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	u := newUsecase(t, mocks.NewMockProfilesServiceClient(ctrl))

	tests := []struct {
		name     string
		password string
		valid    bool
	}{
		{"valid latin", "GoodPass1", true},
		{"valid cyrillic upper + digit", "Хороший1", true},
		{"too short", "G1a", false},
		{"no upper", "goodpass1", false},
		{"no digit", "GoodPass", false},
		{"empty", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := u.validate.Var(tc.password, "required,password")
			if (err == nil) != tc.valid {
				t.Errorf("for %q expected valid=%v, got err=%v", tc.password, tc.valid, err)
			}
		})
	}
}
