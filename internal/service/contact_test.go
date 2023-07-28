package service

import (
	"github.com/harshabangi/bitespeed/internal/storage"
	"github.com/harshabangi/bitespeed/pkg"
	asserts "github.com/stretchr/testify/assert"
	"testing"
)

func Test_ToResponse(t *testing.T) {
	assert := asserts.New(t)

	t.Run("primary contact doesn't have an email", func(t *testing.T) {

		input := []storage.Contact{
			{ID: 1, PhoneNumber: "1", Email: "", LinkPrecedence: primaryContact},
			{ID: 2, PhoneNumber: "1", Email: "a", LinkPrecedence: secondaryContact},
			{ID: 3, PhoneNumber: "2", Email: "a", LinkPrecedence: secondaryContact},
			{ID: 4, PhoneNumber: "2", Email: "b", LinkPrecedence: secondaryContact},
		}

		want := &pkg.ContactResponse{
			Contact: pkg.Contact{
				PrimaryContactID:    1,
				Emails:              []string{"a", "b"},
				PhoneNumbers:        []string{"1", "2"},
				SecondaryContactIDs: []int64{2, 3, 4},
			},
		}

		assert.Equal(want, getResponseFromContacts(input))
	})

	t.Run("primary contact doesn't have a phone number", func(t *testing.T) {

		input := []storage.Contact{
			{ID: 1, PhoneNumber: "", Email: "a", LinkPrecedence: primaryContact},
			{ID: 2, PhoneNumber: "1", Email: "a", LinkPrecedence: secondaryContact},
			{ID: 3, PhoneNumber: "2", Email: "a", LinkPrecedence: secondaryContact},
			{ID: 4, PhoneNumber: "2", Email: "b", LinkPrecedence: secondaryContact},
		}

		want := &pkg.ContactResponse{
			Contact: pkg.Contact{
				PrimaryContactID:    1,
				Emails:              []string{"a", "b"},
				PhoneNumbers:        []string{"1", "2"},
				SecondaryContactIDs: []int64{2, 3, 4},
			},
		}

		assert.Equal(want, getResponseFromContacts(input))
	})

}
