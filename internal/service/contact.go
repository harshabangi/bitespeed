package service

import (
	"github.com/harshabangi/bitespeed/internal/storage"
	"github.com/harshabangi/bitespeed/pkg"
	"github.com/labstack/echo/v4"
	"net/http"
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

	// If either email or phoneNumber or both are not present in any connected component
	// create a new contact and add it as a primary contact
	if len(contacts) == 0 {
		return createContactAndReturnResponse(c, s, req)
	}

	// If either email or phoneNumber is present in the request body
	if req.Email == "" || req.PhoneNumber == "" {
		res, err := getResponse(s.storage, getPrimaryContactID(contacts))
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, res)
	}

	// If both email and phoneNumber is present in the request body
	resp, err := handleContactLinkage(s.storage, req, contacts)
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

func createContactAndReturnResponse(c echo.Context, s *Service, req pkg.ContactRequest) error {
	contact := toContact(req)
	contact.LinkPrecedence = primaryContact

	id, err := s.storage.Contact.CreateContact(contact)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	res := pkg.NewContactResponse().WithID(id)

	if contact.Email != "" {
		res.Contact.Emails = []string{contact.Email}
	}
	if contact.PhoneNumber != "" {
		res.Contact.PhoneNumbers = []string{contact.PhoneNumber}
	}
	return c.JSON(http.StatusOK, &res)
}

func handleContactLinkage(s *storage.Store, req pkg.ContactRequest, contacts []storage.Contact) (*pkg.ContactResponse, error) {

	var contact1, contact2 *storage.Contact

	for i := 0; i < len(contacts); i++ {
		if req.Email == contacts[i].Email {
			contact1 = &contacts[i]
		}
		if req.PhoneNumber == contacts[i].PhoneNumber {
			contact2 = &contacts[i]
		}
	}

	// If only one of email and phone number is new.
	// In that case we will have either email and phone number node in only one connected component.
	// So get create a new contact and derive primary contact id to add it as a linked id for new contact

	if contact1 == nil || contact2 == nil {
		primaryContactID := getPrimaryContactID(contacts)
		c := toContact(req)
		c.LinkedID = primaryContactID
		c.LinkPrecedence = secondaryContact
		if _, err := s.Contact.CreateContact(c); err != nil {
			return nil, err
		}
		return getResponse(s, primaryContactID)
	}

	// If both email and phone number are not new
	switch {
	case contact1.LinkPrecedence == primaryContact && contact2.LinkPrecedence == primaryContact:
		if contact1.ID == contact2.ID {
			return getResponse(s, contact1.ID)
		}
		return linkPrimaryContactsAndGenerateResponse(s, contact1, contact2)

	case contact1.LinkPrecedence == primaryContact && contact2.LinkPrecedence == secondaryContact:
		if contact1.ID == contact2.LinkedID {
			return getResponse(s, contact1.ID)
		}
		c, err := s.Contact.GetContact(contact2.LinkedID)
		if err != nil {
			return nil, err
		}
		return linkPrimaryContactsAndGenerateResponse(s, contact1, c)

	case contact1.LinkPrecedence == secondaryContact && contact2.LinkPrecedence == primaryContact:
		if contact2.ID == contact1.LinkedID {
			return getResponse(s, contact2.ID)
		}
		c, err := s.Contact.GetContact(contact1.LinkedID)
		if err != nil {
			return nil, err
		}
		return linkPrimaryContactsAndGenerateResponse(s, contact2, c)

	case contact1.LinkPrecedence == secondaryContact && contact2.LinkPrecedence == secondaryContact:
		if contact1.LinkedID == contact2.LinkedID {
			return getResponse(s, contact1.LinkedID)
		}
		c1, err := s.Contact.GetContact(contact1.LinkedID)
		if err != nil {
			return nil, err
		}
		c2, err := s.Contact.GetContact(contact2.LinkedID)
		if err != nil {
			return nil, err
		}
		return linkPrimaryContactsAndGenerateResponse(s, c1, c2)

	}

	// shouldn't reach here
	return nil, nil
}

func linkPrimaryContactsAndGenerateResponse(s *storage.Store, primaryContact1, primaryContact2 *storage.Contact) (*pkg.ContactResponse, error) {
	var olderContact, newerContact storage.Contact

	if primaryContact1.CreatedAt.Sub(*primaryContact2.CreatedAt).Seconds() > 0 {
		newerContact = *primaryContact1
		olderContact = *primaryContact2
	} else {
		newerContact = *primaryContact2
		olderContact = *primaryContact1
	}

	if err := s.Contact.UpdateNewerContactsLinkedIDsWithOlderContactsLinkedIDs(olderContact.ID, newerContact.ID); err != nil {
		return nil, err
	}

	if err := s.Contact.UpdateContact(newerContact.ID, storage.Contact{
		LinkedID:       olderContact.ID,
		LinkPrecedence: secondaryContact,
	}); err != nil {
		return nil, err
	}

	return getResponse(s, olderContact.ID)
}

func getPrimaryContactID(contacts []storage.Contact) int64 {
	if contacts[0].LinkPrecedence == primaryContact {
		return contacts[0].ID
	}
	return contacts[0].LinkedID
}

func getResponse(s *storage.Store, primaryContactID int64) (*pkg.ContactResponse, error) {
	allContacts, err := s.Contact.ListContactsByID(primaryContactID)
	if err != nil {
		return nil, err
	}
	return getResponseFromContacts(allContacts), nil
}

func getResponseFromContacts(contacts []storage.Contact) *pkg.ContactResponse {

	var (
		response        = pkg.NewContactResponse()
		emailsMap       = make(map[string]Void)
		phoneNumbersMap = make(map[string]Void)
	)

	for _, c := range contacts {

		if c.LinkPrecedence == primaryContact {
			response.Contact.PrimaryContactID = c.ID
		} else {
			response.Contact.SecondaryContactIDs = append(response.Contact.SecondaryContactIDs, c.ID)
		}

		if c.Email != "" && !keyExists(c.Email, emailsMap) {
			addEmail(c, response)
			emailsMap[c.Email] = VoidValue
		}

		if c.PhoneNumber != "" && !keyExists(c.PhoneNumber, phoneNumbersMap) {
			addPhoneNumber(c, response)
			phoneNumbersMap[c.PhoneNumber] = VoidValue
		}
	}
	return response
}

func addEmail(c storage.Contact, response *pkg.ContactResponse) {
	if c.LinkPrecedence == primaryContact {
		response.Contact.Emails = append([]string{c.Email}, response.Contact.Emails...)
	} else {
		response.Contact.Emails = append(response.Contact.Emails, c.Email)
	}
}

func addPhoneNumber(c storage.Contact, response *pkg.ContactResponse) {
	if c.LinkPrecedence == primaryContact {
		response.Contact.PhoneNumbers = append([]string{c.PhoneNumber}, response.Contact.PhoneNumbers...)
	} else {
		response.Contact.PhoneNumbers = append(response.Contact.PhoneNumbers, c.PhoneNumber)
	}
}
