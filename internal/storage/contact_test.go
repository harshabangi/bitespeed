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
		AddRow(1, "12345", "a@aqfer.com", nil, "primary", &n1).
		AddRow(2, "56789", "a@aqfer.com", 1, "secondary", &n2)

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, phone_number, email, linked_id, link_precedence, created_at FROM contact WHERE email = ? OR phone_number = ?",
	)).
		WithArgs("a@aqfer.com", "12345").
		WillReturnRows(contactRows)

	got, err := s.ListContactsByEmailAndPhoneNumber("a@aqfer.com", "12345")
	assert.Nil(err)

	assert.Equal(2, len(got))
	assert.Equal(Contact{ID: 1, PhoneNumber: "12345", Email: "a@aqfer.com", LinkPrecedence: "primary", CreatedAt: &n1}, got[0])
	assert.Equal(Contact{ID: 2, PhoneNumber: "56789", Email: "a@aqfer.com", LinkedID: 1, LinkPrecedence: "secondary", CreatedAt: &n2}, got[1])
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
		AddRow(2, "56789", "a@aqfer.com", 1, "secondary", &now)

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, phone_number, email, linked_id, link_precedence, created_at FROM contact WHERE linked_id = ? OR id = ?",
	)).WithArgs(2, 2).WillReturnRows(contactRows)

	got, err := s.ListContactsByID(2)
	assert.Nil(err)

	assert.Equal(1, len(got))
	assert.Equal(Contact{ID: 2, PhoneNumber: "56789", Email: "a@aqfer.com", LinkedID: 1, LinkPrecedence: "secondary", CreatedAt: &now}, got[0])
}

func Test_Storage_CreateContact(t *testing.T) {
	assert := asserts.New(t)
	db, mock, err := sqlMock.New()
	assert.Nil(err)

	defer func() { _ = db.Close() }()

	qs := "INSERT contact SET phone_number = ?, email = ?, linked_id = ?, link_precedence = ?"
	mock.ExpectExec(regexp.QuoteMeta(qs)).WithArgs("12345", "a@aqfer.com", 2, "primary").WillReturnResult(sqlMock.NewResult(1, 1))

	s := NewContactStorage(db)
	_, err = s.CreateContact(Contact{PhoneNumber: "12345", Email: "a@aqfer.com", LinkedID: 2, LinkPrecedence: "primary"})
	assert.Nil(err)
}

func Test_Storage_UpdateContact(t *testing.T) {
	assert := asserts.New(t)
	db, mock, err := sqlMock.New()
	assert.Nil(err)

	defer func() { _ = db.Close() }()

	qs := "UPDATE contact SET linked_id = ?, link_precedence = ? WHERE id = ?"
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

	qs := "UPDATE contact SET linked_id = ? WHERE linked_id = ?"
	mock.ExpectExec(regexp.QuoteMeta(qs)).WithArgs(2, 1).WillReturnResult(driver.ResultNoRows)

	s := NewContactStorage(db)
	err = s.UpdateContactsWithNewLinkedIDs(1, 2)
	assert.Nil(err)
}
