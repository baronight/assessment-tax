//go:build !integration
// +build !integration

package db

import (
	"database/sql"
	"log"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/baronight/assessment-tax/models"
)

func NewMock() (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("an error %q was not expected when opening a stub database connection", err)
	}
	return db, mock
}

func TestGetDeductions(t *testing.T) {
	initMock := func() (p Postgres, mock sqlmock.Sqlmock, qry string, rows *sqlmock.Rows) {
		db, mock := NewMock()
		p = Postgres{Db: db}

		qry = "SELECT id, slug, \"name\", amount, \"minAmount\", \"maxAmount\" FROM deductions"

		rows = sqlmock.
			NewRows([]string{"id", "slug", "name", "amount", "minAmount", "maxAmount"})
		return
	}

	t.Run("given success query should return deductions data", func(t *testing.T) {
		p, mock, qry, rows := initMock()
		defer p.Db.Close()

		rows = rows.
			AddRow(1, "k-receipt", "kReceipt", 50000, 0, 100000).
			AddRow(2, "personal", "personalDeduction", 60000, 10000, 100000).
			AddRow(3, "donation", "Donation", 0, 0, 0)
		mock.ExpectQuery(qry).WillReturnRows(rows)

		deductions, err := p.GetDeductions()

		if err != nil {
			t.Errorf("expect no error found but got %q", err)
		}
		want := []models.Deduction{
			{Id: 1, Slug: "k-receipt", Name: "kReceipt", Amount: 50_000, MinAmount: 0, MaxAmount: 100_000},
			{Id: 2, Slug: "personal", Name: "personalDeduction", Amount: 60_000, MinAmount: 10_000, MaxAmount: 100_000},
			{Id: 3, Slug: "donation", Name: "Donation", Amount: 0, MinAmount: 0, MaxAmount: 0},
		}
		if len(want) != len(deductions) {
			t.Errorf("expect deductions have %d rows but got %d rows", len(want), len(deductions))
		}
		if !reflect.DeepEqual(want, deductions) {
			t.Errorf("expect %#v but got %#v", want, deductions)
		}
	})
	t.Run("given no rows found should return no row error", func(t *testing.T) {
		p, mock, qry, _ := initMock()
		defer p.Db.Close()
		mock.ExpectQuery(qry).WillReturnError(sql.ErrNoRows)

		deductions, err := p.GetDeductions()

		if err != sql.ErrNoRows {
			t.Errorf("expect %q but got %q", sql.ErrNoRows, err)
		}
		if deductions != nil {
			t.Errorf("expect deductions should be null, but got %#v", deductions)
		}
	})

	t.Run("given invalid data should return error with null deduction", func(t *testing.T) {
		p, mock, qry, rows := initMock()
		defer p.Db.Close()
		rows = rows.AddRow("1", "slug", "name", "amount", "minAmount", "maxAmount")
		mock.ExpectQuery(qry).WillReturnRows(rows)

		deductions, err := p.GetDeductions()

		if err == nil {
			t.Error("expect error is not nill")
		}
		if deductions != nil {
			t.Errorf("expect deductions should be null, but got %#v", deductions)
		}
	})
}

func TestGetDeduction(t *testing.T) {
	initMock := func() (p Postgres, mock sqlmock.Sqlmock, qry string, rows *sqlmock.Rows) {
		db, mock := NewMock()
		p = Postgres{Db: db}

		qry = "SELECT id, slug, \"name\", amount, \"minAmount\", \"maxAmount\" FROM deductions WHERE slug = \\$1"

		rows = sqlmock.
			NewRows([]string{"id", "slug", "name", "amount", "minAmount", "maxAmount"})
		return
	}

	t.Run("given success query should return deduction", func(t *testing.T) {
		p, mock, qry, rows := initMock()
		defer p.Db.Close()

		rows = rows.
			AddRow(2, "personal", "personalDeduction", 60000, 10000, 100000)

		mock.ExpectQuery(qry).WithArgs("personal").WillReturnRows(rows)

		deductions, err := p.GetDeduction("personal")

		if err != nil {
			t.Errorf("expect no error found but got %q", err)
		}
		want := models.Deduction{
			Id:        2,
			Slug:      "personal",
			Name:      "personalDeduction",
			Amount:    60_000,
			MinAmount: 10_000,
			MaxAmount: 100_000,
		}

		if !reflect.DeepEqual(want, deductions) {
			t.Errorf("expect %#v but got %#v", want, deductions)
		}
	})
	t.Run("given error on query should return error", func(t *testing.T) {
		p, mock, qry, rows := initMock()
		defer p.Db.Close()
		mock.ExpectQuery(qry).WithArgs("personal").WillReturnRows(rows)

		_, err := p.GetDeduction("personal")

		if err == nil {
			t.Errorf("expect error return")
		}
	})
}

func TestUpdateDeduction(t *testing.T) {
	initMock := func() (p Postgres, mock sqlmock.Sqlmock, qry string, rows *sqlmock.Rows) {
		db, mock := NewMock()
		p = Postgres{Db: db}

		qry = "UPDATE deductions SET amount = \\$1 WHERE slug = \\$2" +
			" RETURNING id, slug, \"name\", amount, \"minAmount\", \"maxAmount\""

		rows = sqlmock.
			NewRows([]string{"id", "slug", "name", "amount", "minAmount", "maxAmount"})
		return
	}
	t.Run("given success query should return updated deduction", func(t *testing.T) {
		p, mock, qry, rows := initMock()
		defer p.Db.Close()

		rows = rows.
			AddRow(2, "personal", "personalDeduction", 50000, 10000, 100000)

		mock.ExpectQuery(qry).WithArgs(50000.0, "personal").WillReturnRows(rows)

		deductions, err := p.UpdateDeduction("personal", 50000)

		if err != nil {
			t.Errorf("expect no error found but got %q", err)
		}
		want := models.Deduction{
			Id:        2,
			Slug:      "personal",
			Name:      "personalDeduction",
			Amount:    50_000,
			MinAmount: 10_000,
			MaxAmount: 100_000,
		}

		if !reflect.DeepEqual(want, deductions) {
			t.Errorf("expect %#v but got %#v", want, deductions)
		}
	})
	t.Run("given error on query should return error", func(t *testing.T) {
		p, mock, qry, rows := initMock()
		defer p.Db.Close()
		mock.ExpectQuery(qry).WithArgs(50000, "personal").WillReturnRows(rows)

		_, err := p.UpdateDeduction("personal", 50000)

		if err == nil {
			t.Errorf("expect error return")
		}
	})
}
