package api

import (
	"fmt"
	"github.com/go-chi/render"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	E "github.com/mkuznets/classbox/pkg/api/errors"
	"github.com/mkuznets/classbox/pkg/api/models"
	"github.com/mkuznets/classbox/pkg/api/utils"
	"github.com/mkuznets/classbox/pkg/db"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
)

func (api *API) GetRuns(w http.ResponseWriter, r *http.Request) {

	q := r.URL.Query()

	baselineQ := ""
	if q.Get("baseline") != "" {
		if t, err := strconv.ParseBool(q.Get("baseline")); err == nil {
			baselineQ = fmt.Sprintf("AND r.is_baseline=%v", t)
		} else {
			E.SendError(w, r, nil, http.StatusBadRequest, "`baseline` has to be true or false")
			return
		}
	}

	hashes := &pgtype.TextArray{}
	_ = hashes.Set(q["hash"])

	sql := fmt.Sprintf(`
	SELECT r.hash, r.status, r.output, r.score, t.name, r.is_baseline
	FROM runs as r JOIN tests as t ON (t.id=r.test_id)
	WHERE r.hash=ANY($1) %s
	;`, baselineQ)

	rows, err := api.DB.Query(r.Context(), sql, hashes)
	if err != nil {
		E.Handle(w, r, err)
		return
	}

	runs := make([]*models.Run, 0)

	err = db.IterRows(rows, func(rows pgx.Rows) error {
		run := models.Run{}
		if err := rows.Scan(&run.Hash, &run.Status, &run.Output, &run.Score, &run.Test, &run.Baseline); err != nil {
			return errors.WithStack(err)
		}
		runs = append(runs, &run)
		return nil
	})
	if err != nil {
		E.Handle(w, r, err)
		return
	}

	render.JSON(w, r, &runs)
}

func (api *API) CreateRuns(w http.ResponseWriter, r *http.Request) {

	var runs []models.Run
	if err := render.DecodeJSON(r.Body, &runs); err != nil {
		E.SendError(w, r, err, http.StatusBadRequest, "invalid input")
		return
	}

	testNames := utils.UniqueStrings(runs, "Test")
	testIds, err := api.getTestIds(r.Context(), testNames)
	if err != nil {
		E.Handle(w, r, err)
		return
	}

	err = db.Tx(r.Context(), api.DB, func(tx pgx.Tx) error {
		for _, run := range runs {
			var testID *uint64
			if v, ok := testIds[run.Test]; ok {
				testID = &v
			}
			_, err := tx.Exec(r.Context(), `
			INSERT INTO runs ("hash", status, output, score, test_id, is_baseline)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT ("hash") DO NOTHING
			`, run.Hash, run.Status, run.Output, run.Score, testID, run.Baseline)
			if err != nil {
				return errors.WithStack(err)
			}
		}
		return nil
	})
	if err != nil {
		E.Handle(w, r, err)
		return
	}

	render.NoContent(w, r)
}
