package game

import (
	"encoding/json"
	"fmt"

	"github.com/Popcore/hangmango/pkg/utils"
)

const (
	// MaxWrongChars is the number of failed attempts - or wrong characters -
	// a player can make before the game is considered over.
	MaxWrongChars = 7
)

var (
	// Rules is a help message that descibes the rules of the game and the
	// commands available to players.
	Rules = fmt.Sprintf(`
	GUESS THE CARTOON HERO!

	Rules:
	Guessing the letters in order to discover who the hidden cartoon hero is.
	If you make more that %d mistakes you loose.
	************************************************************************************
	Available commands:
	help             => prints the help screen
	new              => starts a new game
	list             => shows the game history. Each game displays its id and status
	try <character>  => checks if <character> is part of the word to guess
	resume <game-id> => restarts an existing game if its staus is not 'won' or 'game over'
`, MaxWrongChars)
)

// PlayerAction is a custom type that represents the commands a player can
// issues during a game.
type PlayerAction string

const (
	Guess      PlayerAction = "try"
	NewGame    PlayerAction = "new"
	ResumeGame PlayerAction = "resume"
	ListGames  PlayerAction = "list"
	Help       PlayerAction = "help"
	Login      PlayerAction = "login"
)

// State holds information about game status and can be updated according to the
// player's input.
type State struct {
	GameID       int      `json:"id"`
	WordToGuess  string   `json:"word"`
	CharsGuessed []string `json:"guessed"`
	CharsTried   []string `json:"tried"`
	Status       Status   `json:"status"`
}

// Status represents the current status of a game. Its value can be one of the
// three constants defined below.
type Status string

const (
	Unknown    Status = "unknown"
	InProgress Status = "in progress"
	ShowInfo   Status = "show info"
	Paused     Status = "paused"
	GameOver   Status = "game over"
	Won        Status = "won"
	Error      Status = "error"
)

// MarshalJSON is the game State implementation of the JSON Marshaler interface.
// The internal logic formats the word to guess by displaying the characters
// that were guessed and hiding the characther still to guess.
func (g State) MarshalJSON() ([]byte, error) {
	var p string
	for _, c := range g.WordToGuess {
		if utils.Contains(g.CharsGuessed, string(c)) {
			p += fmt.Sprintf("%s ", string(c))
			continue
		}

		p += "_ "
	}

	return json.Marshal(&struct {
		GameID      int      `json:"id"`
		WordToGuess string   `json:"word"`
		CharsTried  []string `json:"tried"`
		Status      Status   `json:"status"`
	}{
		GameID:      g.GameID,
		WordToGuess: p,
		CharsTried:  g.CharsTried,
		Status:      g.Status,
	})
}
