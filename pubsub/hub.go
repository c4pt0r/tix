// Copyright 2022 Ed Huang<i@huangdx.net>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pubsub

import (
	"database/sql"
	"errors"
	"sync"
	"time"

	str2duration "github.com/xhit/go-str2duration/v2"

	"github.com/c4pt0r/log"
	"github.com/c4pt0r/tix"
)

var (
	ErrStreamNotFound error = errors.New("stream not found")
)

type PubSub struct {
	mu          sync.RWMutex
	pollWorkers map[string]*PollWorker
	store       Store
	// streamName -> Stream
	streams map[string]*Stream
	cfg     *tix.Config

	gcWorker   *gcWorker
	gcInterval time.Duration
}

func NewPubSub(c *tix.Config) (*PubSub, error) {
	store, err := OpenStore(c)
	if err != nil {
		return nil, err
	}
	return newPubSubWithStore(store, c), nil
}

func NewPubSubWithDB(db *sql.DB, c *tix.Config) (*PubSub, error) {
	store, err := OpenStoreWithDB(db, c)
	if err != nil {
		return nil, err
	}
	return newPubSubWithStore(store, c), nil
}

func newPubSubWithStore(store Store, c *tix.Config) *PubSub {

	h := &PubSub{
		mu:          sync.RWMutex{},
		cfg:         c,
		store:       store,
		pollWorkers: map[string]*PollWorker{},
		streams:     map[string]*Stream{},
		gcWorker:    newGCWorker(store.DB(), c),
	}
	if c.PubSubConfig.EnableGC {
		gcInterval, err := str2duration.ParseDuration(c.PubSubConfig.GCInterval)
		if err != nil {
			log.Fatal(err)
		}
		h.gcInterval = gcInterval
		go h.gc()
	}
	return h
}

func (m *PubSub) gc() {
	for {
		time.Sleep(m.gcInterval)
		m.mu.RLock()
		// TODO: Should use all stream names(global), not just the ones in the map.
		for streamName := range m.streams {
			m.gcWorker.safeGC(m.cfg.PubSubConfig.StreamTblName(streamName))
		}
		m.mu.RUnlock()
	}
}

func (m *PubSub) ForceGC(streamName string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.gcWorker.safeGC(streamName)
}

func (m *PubSub) Publish(streamName string, msg *Message) error {
	m.mu.Lock()
	if _, ok := m.streams[streamName]; !ok {
		stream, err := NewStream(m.cfg, m.store, streamName)
		if err != nil {
			m.mu.Unlock()
			return err
		}
		if err := stream.Open(); err != nil {
			m.mu.Unlock()
			return err
		}
		m.streams[streamName] = stream
	}
	s := m.streams[streamName]
	m.mu.Unlock()
	s.Publish(msg)
	return nil
}

func (m *PubSub) MinMaxID(streamName string) (int64, int64, error) {
	return m.store.MinMaxID(streamName)
}

func (m *PubSub) PollStat(streamName string) map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if pw, ok := m.pollWorkers[streamName]; ok {
		return pw.Stat()
	}
	return nil
}

func (m *PubSub) MessagesSinceOffset(streamName string, offset Offset) ([]Message, error) {
	var ret []Message
	for {
		log.I("start MessagesSinceOffset", streamName, offset)
		msgs, newOffsetInt, err := m.store.FetchMessages(streamName, offset, m.cfg.MaxTxnSize)
		if err != nil {
			return nil, err
		}
		if len(msgs) > 0 {
			offset = Offset(newOffsetInt)
			ret = append(ret, msgs...)
		} else {
			break
		}
	}
	return ret, nil
}

func (m *PubSub) Subscribe(streamName string, subscriberID string) (<-chan Message, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// if the stream is not in the map, create a new poll worker for this stream
	if _, ok := m.pollWorkers[streamName]; !ok {
		// create a new poll worker for this stream
		pw, err := newPollWorker(m.cfg, m.store, streamName)
		if err != nil {
			return nil, err
		}
		m.pollWorkers[streamName] = pw
	}
	return m.pollWorkers[streamName].addNewSubscriber(subscriberID)
}

func (m *PubSub) Unsubscribe(streamName string, subscriberID string) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if pw, ok := m.pollWorkers[streamName]; ok {
		pw.removeSubscriber(subscriberID)
	}
}

func (m *PubSub) DB() *sql.DB {
	return m.store.DB()
}

func (m *PubSub) GetStreamNames() ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.store.GetStreamNames()
}
