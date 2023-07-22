package service

import (
	"github.com/harshabangi/bitespeed/internal/storage"
	"github.com/harshabangi/bitespeed/pkg"
	"reflect"
	"testing"
)

func Test_ToResponse(t *testing.T) {
	tcc := []struct {
		name  string
		input []storage.Contact
		want  *pkg.ContactResponse
	}{
		{
			"1",
			[]storage.Contact{
				{ID: 1, PhoneNumber: "1", Email: "a", LinkPrecedence: primaryContact},
				{ID: 2, PhoneNumber: "2", Email: "b", LinkPrecedence: secondaryContact},
				{ID: 3, PhoneNumber: "3", Email: "b", LinkPrecedence: secondaryContact},
				{ID: 4, PhoneNumber: "4", Email: "a", LinkPrecedence: secondaryContact},
			},
			&pkg.ContactResponse{
				Contact: pkg.Contact{
					PrimaryContactId:    1,
					Emails:              []string{"a", "b"},
					PhoneNumbers:        []string{"1", "2", "3", "4"},
					SecondaryContactIds: []int64{2, 3, 4},
				},
			},
		},
	}

	for _, tc := range tcc {
		t.Run(tc.name, func(t *testing.T) {
			got := toResponse(tc.input)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %v != want %v", got, tc.want)
			}
		})
	}
}
