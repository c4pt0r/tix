package jobqueue

import (
	"database/sql"
	"sync"
)

type Store interface {
	OpenJobChannel(channelName string) (*JobChannel, error)
}

type TiDBStore struct {
	db  *sql.DB
	cfg *Config

	mu            sync.RWMutex
	mapJobChannel map[string]*JobChannel
}

func OpenStore(cfg *Config) (Store, error) {
	db, err := sql.Open("mysql", cfg.DSN)
	if err != nil {
		return nil, err
	}
	return &TiDBStore{
		db:            db,
		cfg:           cfg,
		mapJobChannel: make(map[string]*JobChannel),
	}, nil
}

func OpenStoreWithDB(db *sql.DB, cfg *Config) Store {
	return &TiDBStore{
		db:            db,
		cfg:           cfg,
		mapJobChannel: make(map[string]*JobChannel),
	}
}

func (s *TiDBStore) Close() {
	s.db.Close()
}

func (s *TiDBStore) DB() *sql.DB {
	return s.db
}

func (s *TiDBStore) GetTablePrefix() string {
	return s.cfg.TablePrefix
}

func (s *TiDBStore) OpenJobChannel(channelName string) (*JobChannel, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.mapJobChannel[channelName]; ok {
		return s.mapJobChannel[channelName], nil
	}

	jc, err := OpenJobChannel(s, channelName)
	if err != nil {
		return nil, err
	}
	s.mapJobChannel[channelName] = jc
	return jc, nil
}
