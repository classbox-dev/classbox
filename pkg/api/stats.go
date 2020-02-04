package api

import (
	"github.com/jackc/pgx/v4"
	"github.com/mkuznets/classbox/pkg/api/models"
	"github.com/mkuznets/classbox/pkg/db"
	"net/http"

	"github.com/go-chi/render"
	E "github.com/mkuznets/classbox/pkg/api/errors"
	"github.com/pkg/errors"
)

func (api *API) GetStats(w http.ResponseWriter, r *http.Request) {

	rows, err := api.DB.Query(r.Context(), `
	SELECT u.login, SUM(s.score) as score, COUNT(*) as count
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
	ORDER BY score DESC;
	`)
	if err != nil {
		E.Handle(w, r, err)
		return
	}

	var stats []*models.UserStat

	err = db.IterRows(rows, func(rows pgx.Rows) error {
		s := models.UserStat{}
		err := rows.Scan(&s.Login, &s.Score, &s.Count)
		if err != nil {
			return errors.WithStack(err)
		}
		stats = append(stats, &s)
		return nil
	})
	if err != nil {
		E.Handle(w, r, err)
	}

	render.JSON(w, r, &stats)
}
