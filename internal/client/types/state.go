package types

import "github.com/google/uuid"

type State struct {
	IsAuthorized bool   `json:"is_authorized"`
	Token        string `json:"token"`
	UserID       int    `json:"user_id"`
	ClientID     string `json:"client_id"`
}

func NewState() *State {
	return &State{
		ClientID: uuid.NewString(),
	}
}
