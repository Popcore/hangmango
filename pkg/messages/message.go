package messages

import (
	"github.com/Popcore/hangmango/pkg/game"
)

// ListGamesResp is the server response type used when a user requires
// the list of games played.
type ListGamesResp struct {
	Games []game.State `json:"games"`
	Error *Error       `json:"error,omitempty"`
}

// GameStateResp is the server response used to desctibe the current game
// state.
type GameStateResp struct {
	State game.State `json:"game"`
	Error *Error     `json:"error,omitempty"`
}

// HelpResp is the server response to a help request. Used to tell the user
// the game rules and the availbale commands.
type HelpResp struct {
	Info  string `json:"info"`
	Error *Error `json:"error,omitempty"`
}

// PlayerReq is the payload send by clients. It must contain the action the
// user wants to perform an an option value.
type PlayerReq struct {
	Action game.PlayerAction `json:"action"`
	Value  string            `json:"value"`
}

type Error struct {
	Message string
}
