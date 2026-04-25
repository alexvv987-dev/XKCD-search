package db

import (
	"context"
	"database/sql"
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
	querySearch = `
	SELECT id, url, words FROM comics WHERE words && $1
	`
	queryGetIds = `
	SELECT id, url, words FROM comics WHERE id = ANY($1)
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

func (db *DB) closeRows(rows *sql.Rows) {
	if err := rows.Close(); err != nil {
		db.log.Error("close rows", "error", err)
	}
}

func (db *DB) GetByIDs(ctx context.Context, ids []int) ([]core.Comics, error) {
	rows, err := db.conn.QueryContext(ctx, queryGetIds, pq.Array(ids))
	if err != nil {
		db.log.Error("get by ids", "error", err)
		return nil, core.ErrInternalServerError
	}
	defer db.closeRows(rows)
	localComics := make([]Comics, 0, len(ids))
	for rows.Next() {
		var comic Comics
		if err = rows.Scan(&comic.ID, &comic.URL, &comic.Words); err != nil {
			db.log.Error("get by ids", "error", err)
			return nil, core.ErrInternalServerError
		}
		localComics = append(localComics, comic)
	}
	comics := make([]core.Comics, 0, len(localComics))
	for _, val := range localComics {
		comics = append(comics, core.Comics{ID: val.ID, URL: val.URL, Words: []string(val.Words)})
	}
	if err = rows.Err(); err != nil {
		db.log.Error("get by ids", "error", err)
		return nil, core.ErrInternalServerError
	}
	return comics, nil
}

func (db *DB) Search(ctx context.Context, words []string) ([]core.Comics, error) {
	localComics := []Comics{}
	if err := db.conn.SelectContext(ctx, &localComics, querySearch, pq.Array(words)); err != nil {
		db.log.Error("search", "error", err)
		return nil, core.ErrInternalServerError
	}
	comics := make([]core.Comics, 0, len(localComics))
	for _, val := range localComics {
		comics = append(comics, core.Comics{ID: val.ID, URL: val.URL, Words: []string(val.Words)})
	}
	return comics, nil
}
