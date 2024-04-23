package db

import (
	"database/sql"
	"os"

	_ "github.com/lib/pq"
)

type Postgres struct {
	Db *sql.DB
}

func New() (*Postgres, error) {
	databaseSource := os.Getenv("DATABASE_URL")
	db, err := sql.Open("postgres", databaseSource)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		defer db.Close()
		return nil, err
	}
	return &Postgres{Db: db}, nil
}
