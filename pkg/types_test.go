package pkg

import (
	asserts "github.com/stretchr/testify/assert"
	"testing"
)

func Test_ValidateContactRequest(t *testing.T) {
	tcc := []struct {
		name      string
		input     ContactRequest
		wantError string
	}{
		{
			"inadequate input parameters. Required either email or phone number or both",
			ContactRequest{},
			"inadequate input parameters. Required either email or phone number or both",
		},
		{
			"incorrect email address: abc",
			ContactRequest{Email: "abc"},
			"incorrect email address: abc",
		},
		{
			"valid request body",
			ContactRequest{Email: "a@gmail.com", PhoneNumber: "12345"},
			"",
		},
	}
	for _, tc := range tcc {
		t.Run(tc.name, func(t *testing.T) {
			assert := asserts.New(t)
			err := tc.input.Validate()

			if tc.wantError == "" {
				assert.Nil(err)
			} else {
				assert.NotNil(err)
			}
		})
	}
}
