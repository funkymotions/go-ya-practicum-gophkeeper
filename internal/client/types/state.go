package types

import (
	"github.com/funkymotions/go-ya-practicum-gophkeeper/internal/model"
	"github.com/google/uuid"
)

type State struct {
	Token    string         `json:"token"`
	UserID   int            `json:"user_id"`
	ClientID string         `json:"client_id"`
	IsOnline bool           `json:"-"`
	Blocks   []*model.Block `json:"-"`
}

func NewState() *State {
	return &State{
		ClientID: uuid.NewString(),
		IsOnline: false,
	}
}
