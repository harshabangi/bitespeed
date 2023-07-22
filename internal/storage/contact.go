package storage

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type ContactStorage interface {
	ListContacts(email string, phoneNumber string) ([]Contact, error)
	CreateContact(contact Contact) (int64, error)
	UpdatedContactsWithNewLinkedIDs(newLinkedID, oldLinkedID int64) error
	UpdatedContact(id int64, linkedID int64) error
	GetContactsByPrimaryID(id int64) ([]Contact, error)
	GetContactByID(id int64) (*Contact, error)
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

func (c *contactStorage) GetContactByID(id int64) (*Contact, error) {

	query := "SELECT phone_number, email, linked_id, link_precedence FROM contact WHERE id = ?"
	row := c.db.QueryRow(query, id)

	var (
		res Contact
		pn  sql.NullString
		e   sql.NullString
		lID sql.NullInt64
	)

	err := row.Scan(&res.ID, &pn, &e, &lID, &res.LinkPrecedence)
	switch err {
	case nil:
		return &res, nil
	case sql.ErrNoRows:
		return nil, nil
	default:
		return nil, err
	}
}

func (c *contactStorage) GetContactsByPrimaryID(id int64) ([]Contact, error) {

	query := "SELECT id, phone_number, email, linked_id, link_precedence, created_at FROM contact WHERE linked_id = ? OR id = ?"

	rows, err := c.db.Query(query, id, id)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = rows.Close()
	}()

	var result []Contact

	for rows.Next() {
		var (
			res Contact
			pn  sql.NullString
			e   sql.NullString
			lID sql.NullInt64
		)

		if err := rows.Scan(&res.ID, &pn, &e, &lID, &res.LinkPrecedence, &res.CreatedAt); err != nil {
			return nil, err
		}

		if pn.Valid {
			res.PhoneNumber = pn.String
		}
		if e.Valid {
			res.Email = e.String
		}
		if lID.Valid {
			res.LinkedID = lID.Int64
		}
		result = append(result, res)
	}
	return result, nil
}

func (c *contactStorage) ListContacts(email string, phoneNumber string) ([]Contact, error) {

	rows, err := c.db.Query("SELECT id, phone_number, email, linked_id, link_precedence, created_at FROM contact WHERE email = ? OR phone_number = ?", email, phoneNumber)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = rows.Close()
	}()

	var result []Contact

	for rows.Next() {
		var (
			res Contact
			pn  sql.NullString
			e   sql.NullString
			lID sql.NullInt64
		)

		if err := rows.Scan(&res.ID, &pn, &e, &lID, &res.LinkPrecedence, &res.CreatedAt); err != nil {
			return nil, err
		}

		if pn.Valid {
			res.PhoneNumber = pn.String
		}
		if e.Valid {
			res.Email = e.String
		}
		if lID.Valid {
			res.LinkedID = lID.Int64
		}

		result = append(result, res)
	}
	return result, rows.Err()
}

func (c *contactStorage) CreateContact(contact Contact) (int64, error) {
	var (
		q  []string
		qp []interface{}
	)

	if contact.Email != "" {
		q = append(q, "email = ?")
		qp = append(qp, contact.Email)
	}

	if contact.PhoneNumber != "" {
		q = append(q, "phone_number = ?")
		qp = append(qp, contact.PhoneNumber)
	}

	if contact.LinkedID != 0 {
		q = append(q, "linked_id = ?")
		qp = append(qp, contact.LinkedID)
	}

	if contact.LinkPrecedence != "" {
		q = append(q, "link_precedence = ?")
		qp = append(qp, contact.LinkPrecedence)
	}

	query := fmt.Sprintf("INSERT INTO contact SET %s", strings.Join(q, ", "))

	r, err := c.db.Exec(query, qp...)
	if err != nil {
		return 0, err
	}
	return r.LastInsertId()
}

func (c *contactStorage) UpdatedContactsWithNewLinkedIDs(newLinkedID, oldLinkedID int64) error {
	_, err := c.db.Exec("UPDATE contact SET linked_id = ? WHERE linked_id = ?", oldLinkedID, newLinkedID)
	return err
}

func (c *contactStorage) UpdatedContact(id int64, linkedID int64) error {
	_, err := c.db.Exec("UPDATE contact SET linked_id = ? WHERE id = ?", linkedID, id)
	return err
}
