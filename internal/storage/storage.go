package storage

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

type database interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

type Store struct {
	Sql     *sql.DB
	Tx      database
	Contact ContactStorage
}

func New(username, password, host, dbname string) (*Store, error) {
	connectString := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", username, password, host, dbname)
	db, err := sql.Open("mysql", connectString)
	if err == nil {
		return &Store{
			Sql:     db,
			Contact: NewContactStorage(db),
		}, nil
	}
	return nil, err
}

func (s *Store) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	tx, err := s.Sql.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	s.Tx = tx
	s.Contact = NewContactStorage(tx)
	return tx, nil
}
