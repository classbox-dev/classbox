package api

import (
	"net/http"
	"sort"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/mkuznets/classbox/pkg/db"

	"github.com/go-chi/render"
	E "github.com/mkuznets/classbox/pkg/api/errors"
	"github.com/pkg/errors"
)

type Solution struct {
	Commit     string    `json:"commit"`
	Test       string    `json:"test"`
	FinishedAt time.Time `json:"finished_at"`
}

func (api *API) GetSolutions(w http.ResponseWriter, r *http.Request) {

	rows, err := api.DB.Query(r.Context(), `
	SELECT
		DISTINCT ON (ci.user_id, ch.test_id)
		u.login, t.finished_at, ci.commit, te.name
	FROM
		checks AS ch
		JOIN commits AS ci ON (ch.commit_id=ci.id)
		JOIN users AS u ON (ci.user_id=u.id)
		JOIN tasks AS t ON (t.commit_id=ci.id)
		JOIN tests as te ON (te.id=ch.test_id)
	WHERE
		ch.status='success' AND ch.is_cached='f' AND ch.name LIKE 'test::%'
	ORDER BY ci.user_id, ch.test_id, ci.id ASC;
	`)
	if err != nil {
		E.Handle(w, r, errors.WithStack(err))
		return
	}

	results := make(map[string][]*Solution)

	err = db.IterRows(rows, func(rows pgx.Rows) error {
		var login string
		s := Solution{}
		err := rows.Scan(&login, &s.FinishedAt, &s.Commit, &s.Test)
		if err != nil {
			return errors.WithStack(err)
		}
		results[login] = append(results[login], &s)
		return nil
	})

	for _, ss := range results {
		sort.Slice(ss, func(i, j int) bool {
			return ss[i].FinishedAt.Before(ss[j].FinishedAt)
		})
	}

	if err != nil {
		E.Handle(w, r, err)
		return
	}

	render.JSON(w, r, &results)
}
