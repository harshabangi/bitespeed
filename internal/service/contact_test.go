package service

import (
	"encoding/json"
	"github.com/harshabangi/bitespeed/internal/storage"
	"github.com/harshabangi/bitespeed/pkg"
	"github.com/labstack/echo/v4"
	asserts "github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func testService(ms *mockContactStorage) *Service {
	return &Service{
		storage: &storage.Store{
			Contact: ms,
		},
	}
}

func Test_ToResponse(t *testing.T) {

	tcc := []struct {
		name  string
		input []storage.Contact
		want  *pkg.ContactResponse
	}{
		{
			"primary contact doesn't have an email",
			[]storage.Contact{
				{ID: 1, PhoneNumber: "1", Email: "", LinkPrecedence: primaryContact},
				{ID: 2, PhoneNumber: "1", Email: "a", LinkPrecedence: secondaryContact},
				{ID: 3, PhoneNumber: "2", Email: "a", LinkPrecedence: secondaryContact},
				{ID: 4, PhoneNumber: "2", Email: "b", LinkPrecedence: secondaryContact},
			},
			&pkg.ContactResponse{
				Contact: pkg.Contact{
					PrimaryContactID:    1,
					Emails:              []string{"a", "b"},
					PhoneNumbers:        []string{"1", "2"},
					SecondaryContactIDs: []int64{2, 3, 4},
				},
			},
		},
		{
			"primary contact doesn't have a phone number",
			[]storage.Contact{
				{ID: 1, PhoneNumber: "", Email: "a", LinkPrecedence: primaryContact},
				{ID: 2, PhoneNumber: "1", Email: "a", LinkPrecedence: secondaryContact},
				{ID: 3, PhoneNumber: "2", Email: "a", LinkPrecedence: secondaryContact},
				{ID: 4, PhoneNumber: "2", Email: "b", LinkPrecedence: secondaryContact},
			},
			&pkg.ContactResponse{
				Contact: pkg.Contact{
					PrimaryContactID:    1,
					Emails:              []string{"a", "b"},
					PhoneNumbers:        []string{"1", "2"},
					SecondaryContactIDs: []int64{2, 3, 4},
				},
			},
		},
	}

	for _, tc := range tcc {
		t.Run(tc.name, func(t *testing.T) {
			assert := asserts.New(t)
			assert.Equal(tc.want, generateContactResponseFromContacts(tc.input))
		})
	}
}

func Test_Identify(t *testing.T) {

	t.Run("create a new contact", func(t *testing.T) {
		assert := asserts.New(t)

		mc := &mockContactStorage{}
		s := testService(mc)

		rqBody := `{"phoneNumber":"12345","email":"a@gmail.com"}`
		req := httptest.NewRequest(http.MethodPost, "/identify", strings.NewReader(rqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := echo.New().NewContext(req, rec)
		c.Set("service", s)

		mc.On("ListContactsByEmailAndPhoneNumber", "a@gmail.com", "12345").Return(([]storage.Contact)(nil), nil)
		mc.On("CreateContact", storage.Contact{Email: "a@gmail.com", PhoneNumber: "12345", LinkPrecedence: primaryContact}).Return(int64(2), nil)

		err := identify(c)
		assert.Nil(err)
		assert.Equal(`{"contact":{"primaryContactId":2,"emails":["a@gmail.com"],"phoneNumbers":["12345"],"secondaryContactIds":[]}}`, strings.Trim(rec.Body.String(), "\n"))

		mc.AssertExpectations(t)
	})

	t.Run("only email is present in the request body", func(t *testing.T) {
		assert := asserts.New(t)

		mc := &mockContactStorage{}
		s := testService(mc)

		rqBody := `{"email":"a@gmail.com"}`
		req := httptest.NewRequest(http.MethodPost, "/identify", strings.NewReader(rqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := echo.New().NewContext(req, rec)
		c.Set("service", s)

		now := time.Now()
		timestamps := []time.Time{now, now.Add(3 * time.Second), now.Add(5 * time.Second)}

		mc.On("ListContactsByEmailAndPhoneNumber", "a@gmail.com", "").Return(
			[]storage.Contact{
				{ID: 2, Email: "a@gmail.com", PhoneNumber: "6789", LinkPrecedence: secondaryContact, LinkedID: 1, CreatedAt: &timestamps[1]},
				{ID: 1, Email: "a@gmail.com", PhoneNumber: "12345", LinkPrecedence: primaryContact, CreatedAt: &timestamps[0]},
			}, nil)

		mc.On("ListContactsByID", int64(1)).Return(
			[]storage.Contact{
				{ID: 1, Email: "a@gmail.com", PhoneNumber: "12345", LinkPrecedence: primaryContact, CreatedAt: &timestamps[0]},
				{ID: 2, Email: "b@gmail.com", PhoneNumber: "12345", LinkPrecedence: secondaryContact, LinkedID: 1, CreatedAt: &timestamps[1]},
				{ID: 3, Email: "b@gmail.com", PhoneNumber: "6789", LinkPrecedence: secondaryContact, LinkedID: 1, CreatedAt: &timestamps[2]},
			}, nil)

		want := &pkg.ContactResponse{
			Contact: pkg.Contact{
				PrimaryContactID:    1,
				Emails:              []string{"a@gmail.com", "b@gmail.com"},
				PhoneNumbers:        []string{"12345", "6789"},
				SecondaryContactIDs: []int64{2, 3},
			},
		}

		wantBytes, _ := json.Marshal(want)
		err := identify(c)
		assert.Nil(err)
		assert.Equal(string(wantBytes), strings.Trim(rec.Body.String(), "\n"))

		mc.AssertExpectations(t)
	})

}
