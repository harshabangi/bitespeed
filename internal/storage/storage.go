package storage

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

type Store struct {
	Sql     *sql.DB
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
