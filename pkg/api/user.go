package api

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/jackc/pgx/v4"
	E "github.com/mkuznets/classbox/pkg/api/errors"
	"github.com/mkuznets/classbox/pkg/api/models"
	"github.com/mkuznets/classbox/pkg/db"
	"github.com/pkg/errors"
)

func (api *API) GetUser(w http.ResponseWriter, r *http.Request) {
	if user, ok := r.Context().Value("User").(*models.User); ok {
		render.JSON(w, r, user)
		return
	}
	E.SendError(w, r, nil, http.StatusUnauthorized, "user not authenticated")
	return
}

func (api *API) GetUserStats(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value("User").(*models.User)
	if !ok {
		E.SendError(w, r, nil, http.StatusUnauthorized, "user not authenticated")
		return
	}

	rows, err := api.DB.Query(r.Context(), `
	SELECT t.name, t.description, t.score, COALESCE(s.passed, 'f')
	FROM tests as t LEFT JOIN (
		SELECT DISTINCT ON (ch.test_id) ch.test_id, (ch.status='success') as passed
		FROM checks as ch
			JOIN commits as ci ON (ci.id=ch.commit_id)
			JOIN tests as t ON (t.id=ch.test_id)
		WHERE ci.user_id=$1 AND ch.test_id IS NOT NULL AND t.is_deleted='f'
		ORDER BY ch.test_id, ch.id DESC
	) as s ON(s.test_id=t.id) WHERE t.is_deleted='f' ORDER BY topic,name;`, user.Id)
	if err != nil {
		E.Handle(w, r, errors.WithStack(err))
		return
	}

	var tests []*models.Test
	err = db.IterRows(rows, func(rows pgx.Rows) error {
		t := models.Test{}
		err := rows.Scan(&t.Name, &t.Description, &t.Score, &t.Passed)
		if err != nil {
			return errors.WithStack(err)
		}
		tests = append(tests, &t)
		return nil
	})
	if err != nil {
		E.Handle(w, r, err)
		return
	}

	stats := &models.UserStats{Tests: tests, Score: 0}
	for _, t := range tests {
		if t.Passed {
			stats.Score += t.Score
		}
		stats.Total += t.Score
	}
	render.JSON(w, r, stats)
}
