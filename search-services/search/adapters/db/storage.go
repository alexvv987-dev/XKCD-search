package db

import (
	"context"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"yadro.com/course/search/core"
)

const (
	queryGet = `
	SELECT id, url, words FROM comics
	`
)

type DB struct {
	log  *slog.Logger
	conn *sqlx.DB
}

type Comics struct {
	ID    int            `db:"id"`
	URL   string         `db:"url"`
	Words pq.StringArray `db:"words"`
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

func (db *DB) Get(ctx context.Context) ([]core.Comics, error) {
	localComics := []Comics{}
	if err := db.conn.SelectContext(ctx, &localComics, queryGet); err != nil {
		db.log.Error("get", "error", err)
		return nil, core.ErrInternalServerError
	}
	comics := make([]core.Comics, 0, len(localComics))
	for _, val := range localComics {
		comics = append(comics, core.Comics{ID: val.ID, URL: val.URL, Words: []string(val.Words)})
	}
	return comics, nil
}
