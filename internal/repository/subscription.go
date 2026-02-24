package repository

import (
	"sync"

	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/model"
)

type subscriptionRepository struct {
	// used map to store subscribers for user apps
	// for permanent state it's better to user in-memory DB
	// like Redis or similar limiting key exp time
	subscribers map[int]map[string]chan []*model.Block
	mu          sync.RWMutex
}

func NewSubscriptionRepository() *subscriptionRepository {
	return &subscriptionRepository{
		subscribers: make(map[int]map[string]chan []*model.Block),
	}
}

func (r *subscriptionRepository) Subscribe(userID int, clientID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	ch, exists := r.subscribers[userID][clientID]
	if !exists {
		ch = make(chan []*model.Block, 1)
		if r.subscribers[userID] == nil {
			r.subscribers[userID] = make(map[string]chan []*model.Block)
		}

		r.subscribers[userID][clientID] = ch
	}
}

func (r *subscriptionRepository) Unsubscribe(userID int, clientID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	ch, exists := r.subscribers[userID][clientID]
	if exists {
		close(ch)
		delete(r.subscribers[userID], clientID)
	}
}

func (r *subscriptionRepository) GetUserSubscribers(userID int) map[string]chan []*model.Block {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.subscribers[userID]
}
