package jobqueue

import (
	"database/sql"
	"sync"
)

// Store is the interface that wraps the basic methods to store jobqueue data.
type Store interface {
	// OpenJobChannel opens a job channel.
	OpenJobChannel(channelName string) (*JobChannel, error)
}

// TiDBStore is the implementation of Store interface.
type TiDBStore struct {
	db  *sql.DB
	cfg *Config
	// mu protect mapJobChannel
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
