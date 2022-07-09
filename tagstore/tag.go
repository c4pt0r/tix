package tagstore

import (
	"errors"
)

var (
	ErrAlreadyElected = errors.New("already elected")
	ErrNotElected     = errors.New("not elected")
)

type TagStore struct {
	cfg *Config
	s   Store
}

type Event struct {
}

func NewTagStore(cfg *Config, campaignName string, candidateName string) (*TagStore, error) {
	s, err := NewStore(cfg)
	if err != nil {
		return nil, err
	}
	c := &TagStore{
		cfg: cfg,
		s:   s,
	}
	return c, nil
}

func (c *TagStore) Init() error {
	// Create tables for tagstore
	return nil
}

func (c *TagStore) Put(key string, tags ...string) error {
	// Put tags to store
	return nil
}

func (c *TagStore) Get(key string) ([]string, error) {
	// Get tags from store
	return nil, nil
}

func (c *TagStore) RemoveKey(key string) error {
	// Remove key from store
	return nil
}

func (c *TagStore) RemoveTag(key string, tag string) error {
	// Remove tag from store
	return nil
}
