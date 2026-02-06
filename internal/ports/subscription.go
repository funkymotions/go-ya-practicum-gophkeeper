package ports

import "github.com/funkymotions/go-ya-practicum-gophkeeper/internal/model"

type SubscriptionService interface {
	Subscribe(userID int, clientID string)
	Unsubscribe(userID int, clientID string)
	NotifySubscribers(userID int, blocks []*model.Block)
	GetUserSubscribers(userID int) map[string]chan []*model.Block
}

type SubscriptionRepository interface {
	Subscribe(userID int, clientID string)
	Unsubscribe(userID int, clientID string)
	GetUserSubscribers(userID int) map[string]chan []*model.Block
}
