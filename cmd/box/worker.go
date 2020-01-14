package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"time"
)

type WebHook struct {
	Login  string `json:"login"`
	Repo   string `json:"repo"`
	Commit string `json:"commit"`
}

func New(ctx context.Context, topic string, db *pgxpool.Pool) *Worker {
	w := Worker{topic: topic, ctx: ctx, db: db}
	go w.do()
	return &w
}

type Worker struct {
	topic string
	ctx   context.Context
	db    *pgxpool.Pool
}

func (w *Worker) do() {
	for {
		err := w.fetchAndProcess()
		if err != nil {
			fmt.Println(err)
		}
		time.Sleep(3 * time.Second)
	}
}

func (w *Worker) fetchAndProcess() error {
	tx, err := w.db.Begin(w.ctx)
	if err != nil {
		return err
	}

	//noinspection GoUnhandledErrorResult
	defer tx.Rollback(w.ctx)

	row := tx.QueryRow(w.ctx, `
		UPDATE tasks SET status='executing', started_at=STATEMENT_TIMESTAMP()
		WHERE id=(
		  SELECT id FROM tasks
		  WHERE status='enqueued' AND topic=$1
		  ORDER BY id
		  FOR UPDATE SKIP LOCKED
		  LIMIT 1
		)
		RETURNING id, payload;
		`, w.topic)

	var (
		tid     uint64
		payload WebHook
	)
	err = row.Scan(&tid, &payload)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			fmt.Println("nothing")
			return nil
		}
		return err
	}

	fmt.Println(tid, payload)

	err = tx.Commit(w.ctx)
	if err != nil {
		return err
	}

	return nil
}
