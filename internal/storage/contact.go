package storage

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type ContactStorage interface {
	ListContactsByEmailAndPhoneNumber(email string, phoneNumber string) ([]Contact, error)
	ListContactsByID(id int64) ([]Contact, error)
	GetContact(id int64) (*Contact, error)
	CreateContact(contact Contact) (int64, error)
	UpdateContact(id int64, contact Contact) error
	UpdateNewerContactsLinkedIDsWithOlderContactsLinkedIDs(olderContactLinkedID, newerContactLinkedID int64) error
}

type contactStorage struct {
	db *sql.DB
}

type Contact struct {
	ID             int64
	PhoneNumber    string
	Email          string
	LinkedID       int64
	LinkPrecedence string
	CreatedAt      *time.Time
	UpdatedAt      *time.Time
	DeletedAt      *time.Time
}

func NewContactStorage(db *sql.DB) ContactStorage {
	return &contactStorage{db: db}
}

func (c *contactStorage) ListContactsByEmailAndPhoneNumber(email string, phoneNumber string) ([]Contact, error) {
	query := "SELECT id, phone_number, email, linked_id, link_precedence, created_at FROM contact WHERE email = ? OR phone_number = ?"

	rows, err := c.db.Query(query, email, phoneNumber)
	if err != nil {
		return nil, err
	}
	return readContacts(rows)
}

func (c *contactStorage) ListContactsByID(id int64) ([]Contact, error) {
	query := "SELECT id, phone_number, email, linked_id, link_precedence, created_at FROM contact WHERE linked_id = ? OR id = ? ORDER BY created_at"

	rows, err := c.db.Query(query, id, id)
	if err != nil {
		return nil, err
	}
	return readContacts(rows)
}

func readContacts(rows *sql.Rows) ([]Contact, error) {
	defer func() {
		_ = rows.Close()
	}()

	var result []Contact

	for rows.Next() {
		var (
			c           Contact
			phoneNumber sql.NullString
			email       sql.NullString
			linkedID    sql.NullInt64
		)

		if err := rows.Scan(&c.ID, &phoneNumber, &email, &linkedID, &c.LinkPrecedence, &c.CreatedAt); err != nil {
			return nil, err
		}

		if phoneNumber.Valid {
			c.PhoneNumber = phoneNumber.String
		}
		if email.Valid {
			c.Email = email.String
		}
		if linkedID.Valid {
			c.LinkedID = linkedID.Int64
		}
		result = append(result, c)
	}
	return result, rows.Err()
}

func (c *contactStorage) GetContact(id int64) (*Contact, error) {
	var (
		query       = "SELECT id, phone_number, email, linked_id, link_precedence, created_at FROM contact WHERE id = ?"
		phoneNumber sql.NullString
		email       sql.NullString
		linkedID    sql.NullInt64

		result Contact
	)

	row := c.db.QueryRow(query, id)

	err := row.Scan(&result.ID, &phoneNumber, &email, &linkedID, &result.LinkPrecedence, &result.CreatedAt)
	if err != nil {
		return nil, err
	}

	if phoneNumber.Valid {
		result.PhoneNumber = phoneNumber.String
	}
	if email.Valid {
		result.Email = email.String
	}
	if linkedID.Valid {
		result.LinkedID = linkedID.Int64
	}

	return &result, nil
}

func (c *contactStorage) CreateContact(contact Contact) (int64, error) {
	var (
		q  []string
		qp []interface{}
	)

	if contact.PhoneNumber != "" {
		q = append(q, "phone_number = ?")
		qp = append(qp, contact.PhoneNumber)
	}

	if contact.Email != "" {
		q = append(q, "email = ?")
		qp = append(qp, contact.Email)
	}

	if contact.LinkedID != 0 {
		q = append(q, "linked_id = ?")
		qp = append(qp, contact.LinkedID)
	}

	if contact.LinkPrecedence != "" {
		q = append(q, "link_precedence = ?")
		qp = append(qp, contact.LinkPrecedence)
	}

	query := fmt.Sprintf("INSERT contact SET %s", strings.Join(q, ", "))

	r, err := c.db.Exec(query, qp...)
	if err != nil {
		return 0, err
	}
	return r.LastInsertId()
}

func (c *contactStorage) UpdateContact(id int64, contact Contact) error {
	var (
		q  []string
		qp []interface{}
	)

	if contact.LinkedID != 0 {
		q = append(q, "linked_id = ?")
		qp = append(qp, contact.LinkedID)
	}

	if contact.LinkPrecedence != "" {
		q = append(q, "link_precedence = ?")
		qp = append(qp, contact.LinkPrecedence)
	}

	query := fmt.Sprintf("UPDATE contact SET %s WHERE id = ?", strings.Join(q, ", "))
	qp = append(qp, id)

	_, err := c.db.Exec(query, qp...)
	return err
}

func (c *contactStorage) UpdateNewerContactsLinkedIDsWithOlderContactsLinkedIDs(olderContactLinkedID, newerContactLinkedID int64) error {
	_, err := c.db.Exec("UPDATE contact SET linked_id = ? WHERE linked_id = ?", olderContactLinkedID, newerContactLinkedID)
	return err
}
