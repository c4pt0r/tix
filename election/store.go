package election

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

type Store interface {
	DB() *sql.DB
}

type TiDBStore struct {
	db  *sql.DB
	cfg *Config
}

func (s *TiDBStore) DB() *sql.DB {
	return s.db
}

func NewStore(cfg *Config) (Store, error) {
	db, err := sql.Open("mysql", cfg.DSN)
	if err != nil {
		return nil, err
	}
	return &TiDBStore{
		cfg: cfg,
		db:  db,
	}, nil
}
