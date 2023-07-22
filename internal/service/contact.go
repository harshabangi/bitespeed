package service

import (
	"github.com/harshabangi/bitespeed/internal/storage"
	"github.com/harshabangi/bitespeed/pkg"
	"github.com/labstack/echo/v4"
	"net/http"
	"sort"
)

const (
	primaryContact   = "primary"
	secondaryContact = "secondary"
)

func identify(c echo.Context) error {
	s := c.Get("service").(*Service)

	var req pkg.ContactRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := req.Validate(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	contacts, err := s.storage.Contact.ListContactsByEmailAndPhoneNumber(req.Email, req.PhoneNumber)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if len(contacts) == 0 {
		contact := toContact(req)
		contact.LinkPrecedence = primaryContact
		id, err := s.storage.Contact.CreateContact(contact)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		return c.JSON(http.StatusOK, &pkg.ContactResponse{Contact: pkg.Contact{
			PrimaryContactId: id,
			Emails:           []string{contact.Email},
			PhoneNumbers:     []string{contact.PhoneNumber},
		}})
	}

	resp, err := helper(s.storage, req, contacts)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, resp)
}

func toContact(rq pkg.ContactRequest) storage.Contact {
	return storage.Contact{
		PhoneNumber: rq.PhoneNumber,
		Email:       rq.Email,
	}
}

func helper(s *storage.Store, req pkg.ContactRequest, contacts []storage.Contact) (*pkg.ContactResponse, error) {

	var primaryContactID int64

	if req.Email == "" || req.PhoneNumber == "" {
		primaryContactID = getPrimaryContactID(contacts)

	} else {

		var (
			e      = 0
			p      = 0
			p1, p2 storage.Contact
		)

		for _, v := range contacts {

			if req.Email == v.Email {
				if v.LinkPrecedence == primaryContact {
					p1 = v
				}
				e++
			}
			if req.PhoneNumber == v.PhoneNumber {
				if v.LinkPrecedence == primaryContact {
					p2 = v
				}
				p++
			}
		}

		if e > 0 && p > 0 {

			if p1.ID != 0 && p2.ID != 0 && p1.ID != p2.ID {

				var cc1, cc2 storage.Contact

				if p1.CreatedAt.Sub(*p2.CreatedAt).Seconds() > 0 {
					cc2 = p1
					cc1 = p2
				} else {
					cc2 = p2
					cc1 = p1
				}

				if err := s.Contact.UpdateContactsWithNewLinkedIDs(cc2.ID, cc1.ID); err != nil {
					return nil, err
				}
				if err := s.Contact.UpdateContact(cc2.ID, storage.Contact{LinkedID: cc1.ID}); err != nil {
					return nil, err
				}

				primaryContactID = cc1.ID
			} else {
				primaryContactID = getPrimaryContactID(contacts)
			}

		} else {
			primaryContactID = getPrimaryContactID(contacts)
			c := toContact(req)
			c.LinkedID = primaryContactID
			c.LinkPrecedence = secondaryContact
			if _, err := s.Contact.CreateContact(c); err != nil {
				return nil, err
			}
		}
	}

	allContacts, err := s.Contact.ListContactsByID(primaryContactID)
	if err != nil {
		return nil, err
	}
	return toResponse(allContacts), nil
}

func getPrimaryContactID(contacts []storage.Contact) int64 {
	if contacts[0].LinkPrecedence == primaryContact {
		return contacts[0].ID
	}
	return contacts[0].LinkedID
}

func toResponse(contacts []storage.Contact) *pkg.ContactResponse {

	result := &pkg.ContactResponse{}
	var primaryContactEmail, primaryContactPhoneNumber string

	emailsMap := make(map[string]struct{})
	phoneNumbersMap := make(map[string]struct{})

	for _, c := range contacts {

		if c.LinkPrecedence == primaryContact {
			result.Contact.PrimaryContactId = c.ID
			primaryContactEmail = c.Email
			primaryContactPhoneNumber = c.PhoneNumber
		} else {
			result.Contact.SecondaryContactIds = append(result.Contact.SecondaryContactIds, c.ID)
		}

		emailsMap[c.Email] = struct{}{}
		phoneNumbersMap[c.PhoneNumber] = struct{}{}
	}

	result.Contact.Emails = append(result.Contact.Emails, primaryContactEmail)
	result.Contact.PhoneNumbers = append(result.Contact.PhoneNumbers, primaryContactPhoneNumber)

	delete(emailsMap, primaryContactEmail)
	delete(phoneNumbersMap, primaryContactPhoneNumber)

	for k := range emailsMap {
		result.Contact.Emails = append(result.Contact.Emails, k)
	}

	for k := range phoneNumbersMap {
		result.Contact.PhoneNumbers = append(result.Contact.PhoneNumbers, k)
	}

	// sorting emails and phone numbers so as it can be testable
	sort.Strings(result.Contact.Emails[1:])
	sort.Strings(result.Contact.PhoneNumbers[1:])

	return result
}
