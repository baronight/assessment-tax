package db

import "github.com/baronight/assessment-tax/models"

// getDeductions implements services.TaxStorer.
func (p *Postgres) GetDeductions() ([]models.Deduction, error) {
	rows, err := p.Db.Query("SELECT id, slug, \"name\", amount, \"minAmount\", \"maxAmount\" FROM deductions")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var deductions []models.Deduction
	for rows.Next() {
		var d models.Deduction
		if err := rows.Scan(
			&d.Id, &d.Slug,
			&d.Name, &d.Amount,
			&d.MinAmount, &d.MaxAmount,
		); err != nil {
			return nil, err
		}
		deductions = append(deductions, d)
	}
	return deductions, nil
}

// GetDeduction implements services.AdminStorer.
func (p *Postgres) GetDeduction(slug string) (models.Deduction, error) {
	row := p.Db.QueryRow("SELECT id, slug, \"name\", amount, \"minAmount\", \"maxAmount\" FROM deductions WHERE slug = $1", slug)
	var deduction models.Deduction
	if err := row.Scan(
		&deduction.Id, &deduction.Slug,
		&deduction.Name, &deduction.Amount,
		&deduction.MinAmount, &deduction.MaxAmount,
	); err != nil {
		return deduction, err
	}
	return deduction, nil
}

// UpdateDeduction implements services.AdminStorer.
func (p *Postgres) UpdateDeduction(slug string, amount float64) (models.Deduction, error) {
	row := p.Db.QueryRow("UPDATE deductions SET amount = $1 WHERE slug = $2"+
		" RETURNING id, slug, \"name\", amount, \"minAmount\", \"maxAmount\"",
		amount, slug)
	var deduction models.Deduction
	if err := row.Scan(
		&deduction.Id, &deduction.Slug,
		&deduction.Name, &deduction.Amount,
		&deduction.MinAmount, &deduction.MaxAmount,
	); err != nil {
		return deduction, err
	}
	return deduction, nil
}
