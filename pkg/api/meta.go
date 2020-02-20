package api

import (
	"fmt"
	"github.com/go-chi/render"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	E "github.com/mkuznets/classbox/pkg/api/errors"
	"github.com/mkuznets/classbox/pkg/api/models"
	"github.com/mkuznets/classbox/pkg/db"
	"github.com/mkuznets/classbox/pkg/utils"
	"github.com/pkg/errors"
	"net/http"
)

func (api *API) GetTests(w http.ResponseWriter, r *http.Request) {

	rows, err := api.DB.Query(r.Context(), `
	SELECT name, description, topic, score FROM tests WHERE is_deleted='f' ORDER BY name
	`)
	if err != nil {
		E.Handle(w, r, errors.WithStack(err))
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

func (api *API) UpdateTests(w http.ResponseWriter, r *http.Request) {

	var tests []models.Test

	if err := render.DecodeJSON(r.Body, &tests); err != nil {
		E.SendError(w, r, err, http.StatusBadRequest, "invalid input")
		return
	}

	for _, test := range tests {
		if test.Name == "" {
			E.SendError(w, r, nil, http.StatusBadRequest, "test name cannot be empty")
			return
		}
	}

	err := db.Tx(r.Context(), api.DB, func(tx pgx.Tx) error {
		for _, test := range tests {
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
		_ = testNames.Set(utils.UniqueStringFields(tests, "Name"))

		_, err := tx.Exec(r.Context(), `UPDATE tests SET is_deleted='t' WHERE NOT (name=ANY($1))`, testNames)
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

func (api *API) GetCourse(w http.ResponseWriter, r *http.Request) {
	var course models.Course
	err := api.DB.QueryRow(r.Context(), `SELECT updated_at, is_ready FROM courses WHERE name='stdlib' LIMIT 1`).Scan(&course.Update, &course.Ready)
	switch {
	case err == pgx.ErrNoRows:
		e := fmt.Errorf("no courses found")
		E.SendError(w, r, e, http.StatusNotFound, e.Error())
		return
	case err != nil:
		E.Handle(w, r, err)
		return
	}
	render.JSON(w, r, &course)
}

func (api *API) UpdateCourse(w http.ResponseWriter, r *http.Request) {

	var course models.Course
	if err := render.DecodeJSON(r.Body, &course); err != nil {
		E.SendError(w, r, err, http.StatusBadRequest, "invalid input")
		return
	}

	err := db.Tx(r.Context(), api.DB, func(tx pgx.Tx) error {
		_, err := tx.Exec(r.Context(), `
		INSERT INTO courses (name, updated_at, is_ready)
		VALUES ('stdlib', STATEMENT_TIMESTAMP(), $1)
		ON CONFLICT ("name") DO UPDATE
		SET name=EXCLUDED.name,
			updated_at=EXCLUDED.updated_at,
			is_ready=EXCLUDED.is_ready
		`, course.Ready)
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
