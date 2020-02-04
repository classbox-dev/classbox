package api

import (
	"github.com/go-chi/render"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	E "github.com/mkuznets/classbox/pkg/api/errors"
	"github.com/mkuznets/classbox/pkg/api/models"
	"github.com/mkuznets/classbox/pkg/api/utils"
	"github.com/mkuznets/classbox/pkg/db"
	"github.com/pkg/errors"
	"net/http"
)

func (api *API) GetMeta(w http.ResponseWriter, r *http.Request) {

	rows, err := api.DB.Query(r.Context(), `
	SELECT name, description, topic, score FROM tests WHERE is_deleted='f' ORDER BY name
	`)
	if err != nil {
		E.Handle(w, r, err)
		return
	}

	var tests []*models.Test

	err = db.IterRows(rows, func(rows pgx.Rows) error {
		t := models.Test{}
		if err := rows.Scan(&t.Name, &t.Description, &t.Topic, &t.Score); err != nil {
			return errors.WithStack(err)
		}
		tests = append(tests, &t)
		return nil
	})

	if err != nil {
		E.Handle(w, r, err)
		return
	}

	render.JSON(w, r, &tests)
}

func (api *API) UpdateMeta(w http.ResponseWriter, r *http.Request) {

	var meta []models.Test

	if err := render.DecodeJSON(r.Body, &meta); err != nil {
		E.SendError(w, r, err, http.StatusBadRequest, "invalid input")
		return
	}

	for _, test := range meta {
		if test.Name == "" {
			E.SendError(w, r, nil, http.StatusBadRequest, "test name cannot be empty")
			return
		}
		if test.Score == 0 {
			E.SendError(w, r, nil, http.StatusBadRequest, "score cannot be zero")
			return
		}
	}

	err := db.Tx(r.Context(), api.DB, func(tx pgx.Tx) error {
		for _, test := range meta {
			_, err := tx.Exec(r.Context(), `
			INSERT INTO tests (name, description, topic, score)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT ("name") DO UPDATE
			SET name=EXCLUDED.name,
				description=EXCLUDED.description,
				topic=EXCLUDED.topic,
				score=EXCLUDED.score,
				is_deleted='f'
			`, test.Name, test.Description, test.Topic, test.Score)
			if err != nil {
				return errors.WithStack(err)
			}
		}

		testNames := &pgtype.TextArray{}
		_ = testNames.Set(utils.UniqueStrings(meta, "Name"))

		_, err := tx.Exec(r.Context(), `UPDATE tests SET is_deleted='t' WHERE name!=ANY($1)`, testNames)
		if err != nil {
			return errors.WithStack(err)
		}

		return nil
	})
	if err != nil {
		E.Handle(w, r, err)
		return
	}

	render.NoContent(w, r)
}
