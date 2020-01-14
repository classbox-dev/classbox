package api

import (
	"net/http"

	"github.com/go-chi/render"
	E "github.com/mkuznets/classbox/pkg/api/errors"
	"github.com/pkg/errors"
)

// ScoreboardRow is a single row in scoreboard table
type scoreboardRow struct {
	FullName string `json:"full_name"`
	Score    int64  `json:"score"`
}

// Scoreboard returns scores of all students
func (api *API) Scoreboard(w http.ResponseWriter, r *http.Request) {

	rows, err := api.DB.Query(r.Context(), `
	SELECT u.full_name as full_name, COALESCE(sb.total, 0) as total
	FROM users as u
	LEFT JOIN (
		SELECT s.user_id, SUM(score) as total
		FROM (
			SELECT DISTINCT ON (s.problem_id, s.user_id) s.user_id, p.score
			FROM submissions as s
			JOIN problems as p
			ON (s.problem_id = p.id)
			WHERE s.is_passed='t'
			ORDER BY s.problem_id, s.user_id, s.id DESC
		) as s
		GROUP BY s.user_id
	) as sb
	ON (u.id=sb.user_id) ORDER BY total DESC, u.id;
	`)
	if err != nil {
		E.Render(w, r, E.Internal(errors.Wrap(err, "DB query failed")))
		return
	}
	//noinspection GoUnhandledErrorResult
	defer rows.Close()

	sbRows := []*scoreboardRow{}

	for rows.Next() {
		sbRow := scoreboardRow{}
		err := rows.Scan(&sbRow.FullName, &sbRow.Score)
		if err != nil {
			E.Render(w, r, E.Internal(errors.Wrap(err, "failed reading DB response")))
			return
		}
		sbRows = append(sbRows, &sbRow)
	}

	if err = rows.Err(); err != nil {
		E.Render(w, r, E.Internal(errors.Wrap(err, "DB query failed")))
		return
	}

	render.JSON(w, r, &sbRows)
}
