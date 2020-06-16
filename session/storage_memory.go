/*
 * Copyright (C) 2018 The "MysteriumNetwork/node" Authors.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package session

import (
	"sync"

	"github.com/mysteriumnetwork/node/eventbus"
	"github.com/mysteriumnetwork/node/identity"
	"github.com/mysteriumnetwork/node/session/event"
)

// NewStorageMemory initiates new session storage
func NewStorageMemory(publisher publisher) *StorageMemory {
	sm := &StorageMemory{
		sessions:  make(map[ID]Session),
		lock:      sync.Mutex{},
		publisher: publisher,
	}
	return sm
}

// StorageMemory maintains all current sessions in memory
type StorageMemory struct {
	sessions  map[ID]Session
	lock      sync.Mutex
	publisher publisher
}

// Add puts given session to storage and publishes a creation event.
// Multiple sessions per peerID is possible in case different services are used
func (storage *StorageMemory) Add(instance Session) {
	storage.lock.Lock()
	defer storage.lock.Unlock()

	storage.sessions[instance.ID] = instance
	go storage.publisher.Publish(event.AppTopicSession, instance.toEvent(event.CreatedStatus))
}

// GetAll returns all sessions in storage
func (storage *StorageMemory) GetAll() []Session {
	storage.lock.Lock()
	defer storage.lock.Unlock()

	// we're never gonna have more than 100000 sessions ongoing on a single node - performance here should not be an issue.
	// see Benchmark_Storage_GetAll
	sessions := make([]Session, len(storage.sessions))

	i := 0
	for _, value := range storage.sessions {
		sessions[i] = value
		i++
	}
	return sessions
}

// Find returns underlying session instance
func (storage *StorageMemory) Find(id ID) (Session, bool) {
	storage.lock.Lock()
	defer storage.lock.Unlock()

	if instance, found := storage.sessions[id]; found {
		if len(storage.sessions) == 1 {
			instance.Last = true
		}
		return instance, true
	}

	return Session{}, false
}

// FindOpts provides fields to search sessions.
type FindOpts struct {
	Peer        *identity.Identity
	ServiceType string
}

// FindBy returns a session by find options.
func (storage *StorageMemory) FindBy(opts FindOpts) (Session, bool) {
	storage.lock.Lock()
	defer storage.lock.Unlock()

	for _, session := range storage.sessions {
		if opts.Peer != nil && *opts.Peer != session.ConsumerID {
			continue
		}
		if opts.ServiceType != "" && opts.ServiceType != session.ServiceType {
			continue
		}
		return session, true
	}
	return Session{}, false
}

// Remove removes given session from underlying storage
func (storage *StorageMemory) Remove(id ID) {
	storage.lock.Lock()
	defer storage.lock.Unlock()

	if instance, found := storage.sessions[id]; found {
		delete(storage.sessions, id)
		go storage.publisher.Publish(event.AppTopicSession, instance.toEvent(event.RemovedStatus))
	}
}

// RemoveForService removes all sessions which belong to given service
func (storage *StorageMemory) RemoveForService(serviceID string) {
	sessions := storage.GetAll()
	for _, session := range sessions {
		if session.ServiceID == serviceID {
			storage.Remove(session.ID)
		}
	}
}

// Subscribe subscribes to relevant events
func (storage *StorageMemory) Subscribe(bus eventbus.Subscriber) error {
	if err := bus.SubscribeAsync(event.AppTopicDataTransferred, storage.consumeDataTransferredEvent); err != nil {
		return err
	}
	if err := bus.SubscribeAsync(event.AppTopicTokensEarned, storage.consumeTokensEarnedEvent); err != nil {
		return err
	}
	return nil
}

// updates the data transfer info on the session
func (storage *StorageMemory) consumeDataTransferredEvent(e event.AppEventDataTransferred) {
	storage.lock.Lock()
	defer storage.lock.Unlock()

	// From a server perspective, bytes up are the actual bytes the client downloaded(aka the bytes we pushed to the consumer)
	// To lessen the confusion, I suggest having the bytes reversed on the session instance.
	// This way, the session will show that it downloaded the bytes in a manner that is easier to comprehend.
	id := ID(e.ID)
	if instance, found := storage.sessions[id]; found {
		instance.DataTransferred.Down = e.Up
		instance.DataTransferred.Up = e.Down
		storage.sessions[id] = instance
	}
}

// updates total tokens earned during the session.
func (storage *StorageMemory) consumeTokensEarnedEvent(e event.AppEventTokensEarned) {
	storage.lock.Lock()
	defer storage.lock.Unlock()

	id := ID(e.SessionID)
	if instance, found := storage.sessions[id]; found {
		instance.TokensEarned = e.Total
		storage.sessions[id] = instance

		go storage.publisher.Publish(event.AppTopicSession, instance.toEvent(event.UpdatedStatus))
	}
}
