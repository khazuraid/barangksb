package handlers

import (
	"context"

	"github.com/jackc/pgx/v5"

	"inventariskantor/internal/models"
)

type ctx2 = context.Context

func collectCategories(rows pgx.Rows) ([]models.Category, error) {
	defer rows.Close()
	var out []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
