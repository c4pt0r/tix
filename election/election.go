package election

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/c4pt0r/log"
)

var (
	ErrAlreadyElected = errors.New("already elected")
	ErrNotElected     = errors.New("not elected")
)

type Campaign struct {
	cfg *Config
	s   Store

	eventChan         chan Event
	campaignName      string
	candidateName     string
	currentLeaderName string
	leaderTerm        int
}

type Event struct {
}

func NewCampaign(cfg *Config, campaignName string, candidateName string) (*Campaign, error) {
	s, err := NewStore(cfg)
	if err != nil {
		return nil, err
	}
	c := &Campaign{
		cfg:           cfg,
		s:             s,
		campaignName:  campaignName,
		candidateName: candidateName,
		eventChan:     make(chan Event),
	}
	return c, nil
}

func (c *Campaign) Init() error {
	stmt := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		name VARCHAR(255) NOT NULL,
		term INTEGER NOT NULL,
		leader VARCHAR(255) NOT NULL,
		lease DATETIME NOT NULL,
		PRIMARY KEY (name)
	)`, c.cfg.TermTable)
	_, err := c.s.DB().Exec(stmt)
	if err != nil {
		return err
	}

	stmt = fmt.Sprintf(`
		INSERT INTO %s (name, term, leader, lease)
		VALUES (?, ?, ?, NOW())
		ON DUPLICATE KEY UPDATE name=name
	`, c.cfg.TermTable)
	_, err = c.s.DB().Exec(stmt, c.campaignName, 0, "")
	if err != nil {
		return err
	}
	return nil
}

// Block until the campaign is elected.
func (c *Campaign) Elect() (<-chan Event, error) {
	for {
		log.I("try to be leader")
		if err := c.tryToBeLeader(); err == nil {
			log.Infof("%s is elected as leader", c.campaignName)
			go func() {
				if err := c.tryToUpdateLease(); err != nil {
					log.Errorf("failed to update lease: %s", err)
				}
				time.Sleep(time.Duration(c.cfg.TermTimeoutInSec/2) * time.Second)
			}()
			log.Infof("elected as leader: %s", c.candidateName)
			return c.eventChan, nil
		}
		time.Sleep(time.Duration(c.cfg.PollIntervalInSec) * time.Second)
	}
}

func (c *Campaign) Resign(ctx context.Context) error {
	stmt := fmt.Sprintf(`
		UPDATE %s SET
			leader = '',
			lease = NOW()
		WHERE
			name = ?
		AND
			leader = ?
		AND
			term = ?
	`, c.cfg.TermTable)
	ret, err := c.s.DB().ExecContext(ctx, stmt, c.campaignName, c.candidateName, c.leaderTerm)
	if err != nil {
		return err
	}
	rows, err := ret.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotElected
	}
	return nil
}

func (c *Campaign) tryToUpdateLease() error {
	if c.currentLeaderName == "" {
		return ErrNotElected
	}

	stmt := fmt.Sprintf(`
		UPDATE %s SET
			lease= NOW() + INTERVAL ? SECOND
		WHERE
			name=? AND leader=? AND term=?
	`, c.cfg.TermTable)
	ret, err := c.s.DB().Exec(stmt, c.cfg.TermTimeoutInSec, c.campaignName, c.currentLeaderName, c.leaderTerm)
	if err != nil {
		return err
	}
	rows, err := ret.RowsAffected()
	if err != nil {
		return err
	}
	// not leader now
	if rows == 0 {
		c.currentLeaderName = ""
		c.leaderTerm = 0
		return ErrNotElected
	}
	return nil
}

func (c *Campaign) tryToBeLeader() error {
	txn, err := c.s.DB().Begin()
	if err != nil {
		return err
	}
	defer txn.Rollback()

	stmt := fmt.Sprintf(`
		SELECT 
			name 
		FROM 
			%s
		WHERE
			name = ? 
		AND 
		(
			leader = ''
				OR
			lease < NOW() 
		)
		FOR UPDATE
	`, c.cfg.TermTable)
	log.D(stmt)
	row := txn.QueryRow(stmt, c.campaignName)
	var campaignName string
	err = row.Scan(&campaignName)
	if err != nil {
		return err
	}
	if campaignName == "" {
		return ErrNotElected
	}

	stmt = fmt.Sprintf(`
		UPDATE %s
		SET
			term = term + 1,
			leader = ?,
			lease = NOW() + INTERVAL ? SECOND
		WHERE
			name = ?
	`, c.cfg.TermTable)
	_, err = txn.Exec(stmt, c.candidateName, c.cfg.TermTimeoutInSec, c.campaignName)
	if err != nil {
		return err
	}

	stmt = fmt.Sprintf(`
		SELECT term FROM %s WHERE name = ?
	`, c.cfg.TermTable)
	row = txn.QueryRow(stmt, c.campaignName)
	var term int
	err = row.Scan(&term)
	if err != nil {
		return err
	}
	err = txn.Commit()
	if err != nil {
		return err
	}
	c.leaderTerm = term
	c.currentLeaderName = c.candidateName
	return nil
}
