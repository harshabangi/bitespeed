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
	db database
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

func NewContactStorage(conn database) ContactStorage {
	return &contactStorage{db: conn}
}

func (c *contactStorage) ListContactsByEmailAndPhoneNumber(email string, phoneNumber string) ([]Contact, error) {
	query := "SELECT id, phone_number, email, linked_id, link_precedence, created_at FROM contact WHERE email = $1 OR phone_number = $2"

	rows, err := c.db.Query(query, email, phoneNumber)
	if err != nil {
		return nil, err
	}
	return readContacts(rows)
}

func (c *contactStorage) ListContactsByID(id int64) ([]Contact, error) {
	query := "SELECT id, phone_number, email, linked_id, link_precedence, created_at FROM contact WHERE linked_id = $1 OR id = $2 ORDER BY created_at"

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
		query       = "SELECT id, phone_number, email, linked_id, link_precedence, created_at FROM contact WHERE id = $1"
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
		ph []string
		qp []interface{}
		i  = 1
	)

	if contact.PhoneNumber != "" {
		q = append(q, "phone_number")
		ph = append(ph, fmt.Sprintf("$%d", i))
		qp = append(qp, contact.PhoneNumber)
		i++
	}

	if contact.Email != "" {
		q = append(q, "email")
		ph = append(ph, fmt.Sprintf("$%d", i))
		qp = append(qp, contact.Email)
		i++
	}

	if contact.LinkedID != 0 {
		q = append(q, "linked_id")
		ph = append(ph, fmt.Sprintf("$%d", i))
		qp = append(qp, contact.LinkedID)
		i++
	}

	if contact.LinkPrecedence != "" {
		q = append(q, "link_precedence")
		ph = append(ph, fmt.Sprintf("$%d", i))
		qp = append(qp, contact.LinkPrecedence)
		i++
	}

	query := fmt.Sprintf("INSERT INTO contact(%s) VALUES(%s) RETURNING id",
		strings.Join(q, ", "), strings.Join(ph, ", "))

	row := c.db.QueryRow(query, qp...)
	var lastInsertID int64
	err := row.Scan(&lastInsertID)
	if err != nil {
		return 0, err
	}
	return lastInsertID, nil
}

func (c *contactStorage) UpdateContact(id int64, contact Contact) error {
	var (
		q  []string
		qp []interface{}
		i  = 1
	)

	if contact.LinkedID != 0 {
		q = append(q, fmt.Sprintf("linked_id = $%d", i))
		qp = append(qp, contact.LinkedID)
		i++
	}

	if contact.LinkPrecedence != "" {
		q = append(q, fmt.Sprintf("link_precedence = $%d", i))
		qp = append(qp, contact.LinkPrecedence)
		i++
	}

	query := fmt.Sprintf("UPDATE contact SET %s WHERE id = $%d", strings.Join(q, ", "), i)
	qp = append(qp, id)

	_, err := c.db.Exec(query, qp...)
	return err
}

func (c *contactStorage) UpdateNewerContactsLinkedIDsWithOlderContactsLinkedIDs(olderContactLinkedID, newerContactLinkedID int64) error {
	_, err := c.db.Exec("UPDATE contact SET linked_id = $1 WHERE linked_id = $2", olderContactLinkedID, newerContactLinkedID)
	return err
}
