package api

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/mkuznets/classbox/pkg/api/models"
	"github.com/mkuznets/classbox/pkg/db"
	"github.com/pkg/errors"
)

func (api *API) getTestIds(ctx context.Context, names []string) (map[string]uint64, error) {
	testNames := &pgtype.TextArray{}
	_ = testNames.Set(names)

	rows, err := api.DB.Query(ctx, `
	SELECT id, name FROM tests WHERE name=ANY($1) AND is_deleted='f'
	`, testNames)
	if err != nil {
		return nil, errors.Wrap(err, "query error")
	}

	testIDs := map[string]uint64{}

	err = db.IterRows(rows, func(rows pgx.Rows) error {
		var (
			id   uint64
			name string
		)
		err := rows.Scan(&id, &name)
		if err != nil {
			return errors.Wrap(err, "failed reading row")
		}
		testIDs[name] = id
		return nil
	})
	if err != nil {
		return nil, err
	}

	return testIDs, nil
}

func (api *API) getRunIds(ctx context.Context, stages []*models.Stage) (map[string]uint64, error) {

	hs := make([]string, 0)
	for _, s := range stages {
		if s.Run != nil {
			hs = append(hs, s.Run.Hash)
		}
	}

	hashes := &pgtype.TextArray{}
	_ = hashes.Set(hs)

	rows, err := api.DB.Query(ctx, `SELECT id, hash FROM runs WHERE hash=ANY($1)`, hashes)
	if err != nil {
		return nil, errors.Wrap(err, "query error")
	}

	runIds := map[string]uint64{}
	err = db.IterRows(rows, func(rows pgx.Rows) error {
		var (
			id   uint64
			hash string
		)
		err := rows.Scan(&id, &hash)
		if err != nil {
			return errors.Wrap(err, "failed reading row")
		}
		runIds[hash] = id
		return nil
	})
	if err != nil {
		return nil, err
	}

	return runIds, nil
}
