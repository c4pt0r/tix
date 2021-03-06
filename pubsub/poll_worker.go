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
	"sync"
	"sync/atomic"
	"time"

	"github.com/c4pt0r/log"
	"github.com/c4pt0r/tix"
	str2duration "github.com/xhit/go-str2duration/v2"
)

// PollWorker is a worker that polls messages from a stream
type PollWorker struct {
	cfg            *tix.Config
	streamName     string
	lastSeenOffset Offset
	store          Store
	stopped        atomic.Value
	numSubscribers int32
	pollInterval   time.Duration

	// make sure subscribers is threadsafe here
	mu sync.Mutex
	// subscribers map[string]Subscriber, key is subscriber id
	subscribers map[string]chan Message
}

func newPollWorker(cfg *tix.Config, s Store, streamName string) (*PollWorker, error) {
	// create stream table
	err := s.CreateStream(streamName)
	if err != nil {
		return nil, err
	}

	// get last seen offset
	_, offset, err := s.MinMaxID(streamName)
	if err != nil {
		return nil, err
	}

	stopped := atomic.Value{}
	stopped.Store(false)

	pollInterval, err := str2duration.ParseDuration(cfg.PubSubConfig.PollInterval)
	if err != nil {
		return nil, err
	}

	pw := &PollWorker{
		streamName:     streamName,
		cfg:            cfg,
		lastSeenOffset: Offset(offset),
		store:          s,
		stopped:        stopped,
		numSubscribers: 0,
		mu:             sync.Mutex{},
		subscribers:    map[string]chan Message{},
		pollInterval:   pollInterval,
	}

	go pw.run()
	return pw, nil
}

func (pw *PollWorker) addNewSubscriber(subscriberID string) (<-chan Message, error) {
	pw.mu.Lock()
	defer pw.mu.Unlock()
	log.I("pollWorkers", pw.streamName, "got new subscriber:", subscriberID)
	ch := make(chan Message)
	/*
		if offset != LatestId && offset < pw.lastSeenOffset {
			for {
				msgs, newOffsetInt, err := pw.store.FetchMessages(pw.streamName, offset, pw.cfg.MaxBatchSize)
				if err != nil {
					return nil, err
				}
				if len(msgs) > 0 {
					offset = Offset(newOffsetInt)
					go func() {
						// send messages to this subscriber
						for _, msg := range msgs {
							ch <- msg
						}
					}()
				} else {
					break
				}
			}
		}
	*/
	// listen to new messages
	pw.subscribers[subscriberID] = ch
	atomic.AddInt32(&pw.numSubscribers, 1)
	return ch, nil
}

func (pw *PollWorker) Stat() map[string]interface{} {
	return map[string]interface{}{
		"last_poll_id":        pw.lastSeenOffset,
		"poll_interval_in_ms": pw.cfg.PubSubConfig.PollInterval,
		"poll_batch_size":     pw.cfg.MaxTxnSize,
	}
}

func (pw *PollWorker) removeSubscriber(subscriberID string) {
	pw.mu.Lock()
	defer pw.mu.Unlock()
	log.I("pollWorkers", pw.streamName, "remove subscriber:", subscriberID)
	if v, ok := pw.subscribers[subscriberID]; ok {
		close(v)
	}
	delete(pw.subscribers, subscriberID)
	atomic.AddInt32(&pw.numSubscribers, -1)
}

func (pw *PollWorker) Stop() {
	log.I("pollWorkers", pw.streamName, "stopped")
	pw.stopped.Store(true)
}

func (pw *PollWorker) run() {
	log.Info("sub: start polling from", pw.streamName, "@id=", pw.lastSeenOffset)
	for !pw.stopped.Load().(bool) {
		// get messages from the stream in batches
		msgs, max, err := pw.store.FetchMessages(pw.streamName, pw.lastSeenOffset, pw.cfg.MaxTxnSize)
		if err != nil {
			log.Error(err)
			time.Sleep(pw.pollInterval)
			goto done
		}
		if len(msgs) > 0 {
			pw.lastSeenOffset = Offset(max)
			log.Info("sub: got", len(msgs), "messages from", pw.streamName, "@ id=", pw.lastSeenOffset)

			pw.mu.Lock()
			// fanout to subscribers
			for _, c := range pw.subscribers {
				go func(ch chan Message) {
					for _, msg := range msgs {
						ch <- msg
					}
				}(c)
			}
			pw.mu.Unlock()
		}
	done:
		time.Sleep(pw.pollInterval)
	}
	log.D("poll worker stopped")
}
