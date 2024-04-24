package db

import "github.com/baronight/assessment-tax/models"

// getDeductions implements services.TaxStorer.
func (p *Postgres) GetDeductions() ([]models.Deduction, error) {
	rows, err := p.Db.Query("SELECT id, slug, name, amount, minAmount, maxAmount FROM deductions")
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
