package service

import (
	"fmt"

	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/model"
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/ports"
)

type subscriptionService struct {
	subscriptionRepository ports.SubscriptionRepository
}

var _ ports.SubscriptionService = (*subscriptionService)(nil)

func NewSubscriptionService(subscriptionRepository ports.SubscriptionRepository) *subscriptionService {
	return &subscriptionService{
		subscriptionRepository: subscriptionRepository,
	}
}

func (s *subscriptionService) Subscribe(userID int, clientID string) {
	s.subscriptionRepository.Subscribe(userID, clientID)
}

func (s *subscriptionService) Unsubscribe(userID int, clientID string) {
	s.subscriptionRepository.Unsubscribe(userID, clientID)
}

func (s *subscriptionService) NotifySubscribers(userID int, blocks []*model.Block) {
	userClients := s.subscriptionRepository.GetUserSubscribers(userID)
	for clientID, ch := range userClients {
		select {
		case ch <- blocks:
		default:
			// If the channel is full, skip sending to avoid blocking
			fmt.Printf("Skipping subscriber %s: channel is full\n", clientID)
		}
	}
}

func (s *subscriptionService) GetUserSubscribers(userID int) map[string]chan []*model.Block {
	return s.subscriptionRepository.GetUserSubscribers(userID)
}
