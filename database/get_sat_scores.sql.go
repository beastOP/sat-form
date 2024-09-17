// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: get_sat_scores.sql

package database

import (
	"context"
)

const getSATScores = `-- name: GetSATScores :many
SELECT id, name, address, city, country, pincode, sat_score, passed, created_at, updated_at FROM sat_scores
`

func (q *Queries) GetSATScores(ctx context.Context) ([]SatScore, error) {
	rows, err := q.db.QueryContext(ctx, getSATScores)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SatScore
	for rows.Next() {
		var i SatScore
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Address,
			&i.City,
			&i.Country,
			&i.Pincode,
			&i.SatScore,
			&i.Passed,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
