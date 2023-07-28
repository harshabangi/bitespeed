package pkg

import (
	"fmt"
	"net/mail"
	"regexp"
)

type ContactRequest struct {
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
}

func (c *ContactRequest) Validate() error {
	if c.Email == "" && c.PhoneNumber == "" {
		return fmt.Errorf("inadequate input parameters. Required either email or phone number or both")
	}
	if err := validateEmail(c.Email); err != nil {
		return err
	}
	return nil
}

func validateEmail(email string) error {
	if email == "" {
		return nil
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf("incorrect email address: %s", email)
	}
	return nil
}

func validatePhoneNumber(phoneNumber string) error {
	if phoneNumber == "" {
		return nil
	}
	reg := regexp.MustCompile(`^\+?\d{1,3}?[-.\s]?\(?\d{1,3}\)?[-.\s]?\d{1,4}[-.\s]?\d{1,4}[-.\s]?\d{1,9}$`)

	if reg.MatchString(phoneNumber) {
		return nil
	}

	return fmt.Errorf("incorrect phone number: %s", phoneNumber)
}

type ContactResponse struct {
	Contact Contact `json:"contact"`
}

func NewContactResponse() *ContactResponse {
	return &ContactResponse{
		Contact: Contact{
			Emails:              make([]string, 0),
			PhoneNumbers:        make([]string, 0),
			SecondaryContactIDs: make([]int64, 0),
		},
	}
}

func (c *ContactResponse) WithID(id int64) *ContactResponse {
	c.Contact.PrimaryContactID = id
	return c
}

type Contact struct {
	PrimaryContactID    int64    `json:"primaryContactId"`
	Emails              []string `json:"emails"`
	PhoneNumbers        []string `json:"phoneNumbers"`
	SecondaryContactIDs []int64  `json:"secondaryContactIds"`
}
