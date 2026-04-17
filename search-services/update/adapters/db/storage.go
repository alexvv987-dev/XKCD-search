package db

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"yadro.com/course/update/core"
)

const (
	queryAdd = `
	INSERT INTO comics (id, url, words) VALUES ($1, $2, $3)
	`
	queryStats = `
    SELECT
		(SELECT COUNT(*) FROM comics) as comics_fetched,
		COALESCE(SUM(array_length(words, 1)), 0) as words_total,
		COUNT(DISTINCT word) as words_unique
	FROM comics, unnest(words) as word
	`
	queryIDs = `
	SELECT id FROM comics
	`
	queryDrop = `
	TRUNCATE TABLE comics
`
)

type DB struct {
	log  *slog.Logger
	conn *sqlx.DB
}

func New(log *slog.Logger, address string) (*DB, error) {

	db, err := sqlx.Connect("pgx", address)
	if err != nil {
		log.Error("connection problem", "address", address, "error", err)
		return nil, core.ErrInternalServerError
	}
	return &DB{
		log:  log,
		conn: db,
	}, nil
}

func (db *DB) Add(ctx context.Context, comics core.Comics) error {
	_, err := db.conn.ExecContext(ctx, queryAdd, comics.ID, comics.URL, pq.Array(comics.Words))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return core.ErrAlreadyExists
		}
		db.log.Error("insert", "error", err)
		return err
	}
	return nil
}

func (db *DB) Stats(ctx context.Context) (core.DBStats, error) {
	row := db.conn.QueryRowxContext(ctx, queryStats)
	stats := core.DBStats{}
	if err := row.Scan(&stats.ComicsFetched, &stats.WordsTotal, &stats.WordsUnique); err != nil {
		db.log.Error("stats", "error", err)
		return core.DBStats{}, core.ErrInternalServerError
	}
	return stats, nil
}

func (db *DB) IDs(ctx context.Context) ([]int, error) {
	var ids []int
	if err := db.conn.SelectContext(ctx, &ids, queryIDs); err != nil {
		db.log.Error("ids", "error", err)
		return nil, core.ErrInternalServerError
	}
	return ids, nil
}

func (db *DB) Drop(ctx context.Context) error {
	if _, err := db.conn.ExecContext(ctx, queryDrop); err != nil {
		db.log.Error("drop", "error", err)
		return core.ErrInternalServerError
	}
	return nil
}
