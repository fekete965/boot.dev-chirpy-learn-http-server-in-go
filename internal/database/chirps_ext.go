package database

import (
	"context"
	"strings"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/constants"
	"github.com/google/uuid"
)

type GetAllChirpsParams struct {
	AuthorID *uuid.UUID
	Sort *string
}

func getSortQuery(sort *string, defaultSort *string) string {
	sortValue := sort

	if sortValue == nil {
		sortValue = defaultSort
	}

	if sortValue == nil {
		return ""
	}

	switch *sortValue {
	case "desc":
		return "ORDER BY created_at DESC;"
	default:
		return "ORDER BY created_at ASC;"
	}
}

func (q *Queries) GetAllChirps(ctx context.Context, arg GetAllChirpsParams) ([]Chirp, error) {
	var args []any

	querySegments := []string{"SELECT * FROM chirps"}
	if arg.AuthorID != nil {
		querySegments = append(querySegments, "WHERE user_id = $1")
		args = append(args, *arg.AuthorID)
	}

	sortQuery := getSortQuery(arg.Sort, &constants.DEFAULT_SORT)
	querySegments = append(querySegments, sortQuery)

	rows, err := q.db.QueryContext(ctx, strings.Join(querySegments, " "), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var items []Chirp

	for rows.Next() {
		var i Chirp
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Body,
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
