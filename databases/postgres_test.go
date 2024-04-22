package databases

import (
	"os"
	"testing"
)

const (
	VALID_DB_ENV   string = "host=localhost port=5433 user=postgres password=postgres dbname=ktaxes sslmode=disable"
	INVALID_DB_ENV string = "helloworld"
)

func TestDB(t *testing.T) {
	t.Run("given wrong env should return error", func(t *testing.T) {
		os.Setenv("DATABASE_URL", INVALID_DB_ENV)
		_, err := New()

		if err == nil {
			t.Errorf("expect error was not nil")
		}
	})

	t.Run("given valid env should return Postgres which can execution sql", func(t *testing.T) {
		os.Setenv("DATABASE_URL", VALID_DB_ENV)
		pg, err := New()

		if err != nil {
			t.Fatal("expect error was nil")
		}
		defer pg.Db.Close()

		row := pg.Db.QueryRow("SELECT 1 + 1")
		var expect int
		if err := row.Scan(&expect); err != nil {
			t.Error("expect error was nil")
		}
		if expect != 2 {
			t.Errorf("expect result should be 2 but got %d", expect)
		}
	})
}
