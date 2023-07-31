package storage

import (
	"database/sql/driver"
	sqlMock "github.com/DATA-DOG/go-sqlmock"
	asserts "github.com/stretchr/testify/assert"
	"regexp"
	"testing"
	"time"
)

func Test_Storage_ListContactsByEmailAndPhoneNumber(t *testing.T) {
	assert := asserts.New(t)
	db, mock, err := sqlMock.New()
	assert.Nil(err)

	defer func() {
		_ = db.Close()
	}()

	s := NewContactStorage(db)

	n1 := time.Now().UTC()
	n2 := n1.Add(40 * time.Second)

	contactRows := sqlMock.NewRows([]string{"id", "phone_number", "email", "linked_id", "link_precedence", "created_at"}).
		AddRow(1, "12345", "a@gmail.com", nil, "primary", &n1).
		AddRow(2, "56789", "a@gmail.com", 1, "secondary", &n2)

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, phone_number, email, linked_id, link_precedence, created_at FROM contact WHERE email = $1 OR phone_number = $2",
	)).
		WithArgs("a@gmail.com", "12345").
		WillReturnRows(contactRows)

	got, err := s.ListContactsByEmailAndPhoneNumber("a@gmail.com", "12345")
	assert.Nil(err)

	assert.Equal(2, len(got))
	assert.Equal(Contact{ID: 1, PhoneNumber: "12345", Email: "a@gmail.com", LinkPrecedence: "primary", CreatedAt: &n1}, got[0])
	assert.Equal(Contact{ID: 2, PhoneNumber: "56789", Email: "a@gmail.com", LinkedID: 1, LinkPrecedence: "secondary", CreatedAt: &n2}, got[1])
}

func Test_Storage_ListContactsByID(t *testing.T) {
	assert := asserts.New(t)
	db, mock, err := sqlMock.New()
	assert.Nil(err)

	defer func() {
		_ = db.Close()
	}()

	s := NewContactStorage(db)

	now := time.Now().UTC()

	contactRows := sqlMock.NewRows([]string{"id", "phone_number", "email", "linked_id", "link_precedence", "created_at"}).
		AddRow(2, "56789", "a@gmail.com", 1, "secondary", &now)

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, phone_number, email, linked_id, link_precedence, created_at FROM contact WHERE linked_id = $1 OR id = $2",
	)).WithArgs(2, 2).WillReturnRows(contactRows)

	got, err := s.ListContactsByID(2)
	assert.Nil(err)

	assert.Equal(1, len(got))
	assert.Equal(Contact{ID: 2, PhoneNumber: "56789", Email: "a@gmail.com", LinkedID: 1, LinkPrecedence: "secondary", CreatedAt: &now}, got[0])
}

func Test_Storage_GetContact(t *testing.T) {
	assert := asserts.New(t)
	db, mock, err := sqlMock.New()
	assert.Nil(err)

	defer func() { _ = db.Close() }()

	now := time.Now().UTC()
	contactRows := sqlMock.NewRows([]string{"id", "phone_number", "email", "linked_id", "link_precedence", "created_at"}).
		AddRow(2, "56789", "a@gmail.com", 1, "secondary", &now)

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, phone_number, email, linked_id, link_precedence, created_at FROM contact WHERE id = $1",
	)).WithArgs(2).WillReturnRows(contactRows)

	s := NewContactStorage(db)
	got, err := s.GetContact(2)
	assert.Nil(err)
	assert.Equal(&Contact{
		ID:             2,
		PhoneNumber:    "56789",
		Email:          "a@gmail.com",
		LinkedID:       1,
		LinkPrecedence: "secondary",
		CreatedAt:      &now,
	}, got)
}

func Test_Storage_CreateContact(t *testing.T) {
	assert := asserts.New(t)
	db, mock, err := sqlMock.New()
	assert.Nil(err)

	defer func() { _ = db.Close() }()

	rows := sqlMock.NewRows([]string{"id"}).AddRow(1)

	qs := "INSERT INTO contact(phone_number, email, linked_id, link_precedence) VALUES($1, $2, $3, $4) RETURNING id"
	mock.ExpectQuery(regexp.QuoteMeta(qs)).WithArgs("12345", "a@gmail.com", 2, "primary").WillReturnRows(rows)

	s := NewContactStorage(db)
	_, err = s.CreateContact(Contact{PhoneNumber: "12345", Email: "a@gmail.com", LinkedID: 2, LinkPrecedence: "primary"})
	assert.Nil(err)
}

func Test_Storage_UpdateContact(t *testing.T) {
	assert := asserts.New(t)
	db, mock, err := sqlMock.New()
	assert.Nil(err)

	defer func() { _ = db.Close() }()

	qs := "UPDATE contact SET linked_id = $1, link_precedence = $2 WHERE id = $3"
	mock.ExpectExec(regexp.QuoteMeta(qs)).WithArgs(2, "primary", 1).WillReturnResult(driver.ResultNoRows)

	s := NewContactStorage(db)
	err = s.UpdateContact(1, Contact{LinkedID: 2, LinkPrecedence: "primary"})
	assert.Nil(err)
}

func Test_Storage_UpdateContactsWithNewLinkedIDs(t *testing.T) {
	assert := asserts.New(t)
	db, mock, err := sqlMock.New()
	assert.Nil(err)

	defer func() { _ = db.Close() }()

	qs := "UPDATE contact SET linked_id = $1 WHERE linked_id = $2"
	mock.ExpectExec(regexp.QuoteMeta(qs)).WithArgs(1, 2).WillReturnResult(driver.ResultNoRows)

	s := NewContactStorage(db)
	err = s.UpdateNewerContactsLinkedIDsWithOlderContactsLinkedIDs(1, 2)
	assert.Nil(err)
}
