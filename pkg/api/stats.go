package api

import (
	"net/http"

	"github.com/jackc/pgx/v4"
	"github.com/mkuznets/classbox/pkg/api/models"
	"github.com/mkuznets/classbox/pkg/db"

	"github.com/go-chi/render"
	E "github.com/mkuznets/classbox/pkg/api/errors"
	"github.com/pkg/errors"
)

func (api *API) GetStats(w http.ResponseWriter, r *http.Request) {

	rows, err := api.DB.Query(r.Context(), `
	SELECT u.login, COALESCE(st.score, 0) as score, COALESCE(st.count, 0) as count
	FROM users as u LEFT JOIN (
		SELECT u.id as user_id, SUM(s.score) as score, COUNT(*) as count
		FROM (
			SELECT DISTINCT ON (ch.test_id, ci.user_id) ci.user_id, t.score, ch.status
			FROM checks as ch
				JOIN commits as ci ON (ci.id=ch.commit_id)
				JOIN tests as t ON (t.id=ch.test_id)
			WHERE ch.test_id IS NOT NULL AND t.is_deleted='f'
			ORDER BY ch.test_id, ci.user_id, ch.id DESC
		) as s
		JOIN users as u ON (u.id=s.user_id)
		WHERE s.status='success' GROUP BY u.id
	) as st ON (u.id=st.user_id)
	ORDER BY score DESC, login;
	`)
	if err != nil {
		E.Handle(w, r, errors.WithStack(err))
		return
	}

	var stats []*models.Stat

	err = db.IterRows(rows, func(rows pgx.Rows) error {
		s := models.Stat{}
		err := rows.Scan(&s.Login, &s.Score, &s.Count)
		if err != nil {
			return errors.WithStack(err)
		}
		stats = append(stats, &s)
		return nil
	})
	if err != nil {
		E.Handle(w, r, err)
		return
	}

	render.JSON(w, r, &stats)
}
