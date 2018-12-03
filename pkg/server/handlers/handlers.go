package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/Popcore/hangmango/pkg/game"
	"github.com/Popcore/hangmango/pkg/messages"
	"github.com/Popcore/hangmango/pkg/store"
	"github.com/Popcore/hangmango/pkg/utils"
)

// System holds services and configuration settings required by the game controller.
type System struct {
	Logger *log.Logger
	Store  store.Storer
}

// controller holds all the required information in order to manage game sessions
// for a connected user.
type controller struct {
	Conn      net.Conn
	System    System
	UserID    string
	GameState *game.State
	Encoder   *json.Encoder
}

// NewSession returns a controller instance that cen be used to manage games.
// It is the actual game entry point.
func NewSession(System System, conn net.Conn) error {
	h := &controller{
		System:  System,
		Conn:    conn,
		Encoder: json.NewEncoder(conn),
	}

	return h.handleGameIO()
}

// handleGameIO glues together the input parsing, processing and response processes.
func (c *controller) handleGameIO() (err error) {

	for {
		cmd, err := parseUserInput(c.Conn)
		if err != nil {
			if err == io.EOF {
				c.System.Logger.Printf("user %s disconnected", c.UserID)
				break
			}

			c.System.Logger.Println(err)
			c.Encoder.Encode(err)
		}

		err = c.handlePlayerAction(*cmd)
		if err != nil {
			if err == io.EOF {
				c.System.Logger.Printf("user %s disconnected", c.UserID)
				break
			}
			c.System.Logger.Println(err)
			c.Encoder.Encode(err)
		}
	}

	return err
}

// loginHandler sets the controller UserID using the name received from
// the user.
func (c *controller) loginHandler(userName string) error {
	c.System.Logger.Printf("user authenticated: %s", userName)

	err := c.System.Store.SaveNewUser(userName)
	if err != nil {
		return c.Encoder.Encode(messages.HelpResp{
			Error: &messages.Error{Message: err.Error()},
		})
	}

	c.UserID = userName

	return c.Encoder.Encode(messages.HelpResp{Info: game.Rules})
}

// newGameHandler returns a new game as saves the previous game if is not nil.
func (c *controller) newGameHandler() error {
	c.System.Logger.Printf("%s is starting a new game", c.UserID)

	if c.GameState != nil {
		c.GameState.Status = game.Paused

		_, err := c.System.Store.SaveGame(c.UserID, *c.GameState)
		if err != nil {
			return err
		}
	}

	newGame := game.State{
		WordToGuess: randomHero(),
		CharsTried:  []string{},
		Status:      game.InProgress,
	}
	saved, err := c.System.Store.SaveGame(c.UserID, newGame)
	if err != nil {
		return err
	}

	c.GameState = saved
	c.GameState.Status = game.InProgress

	return c.Encoder.Encode(messages.GameStateResp{State: *c.GameState})
}

// helpHandler returns the game rules and available commands to interact with
// the server.
func (c *controller) helpHandler() error {
	c.System.Logger.Printf("%s is requesting help", c.UserID)

	return c.Encoder.Encode(messages.HelpResp{
		Info: game.Rules,
	})
}

// listGamesHandler returns a list of existing games that belong to the current
// user or an empty slice if no games are found.
func (c *controller) listGamesHandler() error {
	c.System.Logger.Printf("%s is listing games played", c.UserID)

	if c.GameState != nil && c.GameState.Status == game.InProgress {
		c.GameState.Status = game.Paused

		saved, err := c.System.Store.SaveGame(c.UserID, *c.GameState)
		if err != nil {
			return err
		}

		c.GameState = saved
	}

	games, err := c.System.Store.GetGamesByUser(c.UserID)
	if err != nil {
		return err
	}

	return c.Encoder.Encode(messages.ListGamesResp{
		Games: games,
	})
}

// resumeGameHandler sets the game identified by the gameID as the current game.
// It save the existing game if in progress.
func (c *controller) resumeGameHandler(gameID string) error {
	c.System.Logger.Printf("%s is resuming game %s", c.UserID, gameID)

	id, err := strconv.Atoi(gameID)
	if err != nil {
		return err
	}

	toResume, err := c.System.Store.GetGameByID(c.UserID, id)
	if err != nil {
		return c.Encoder.Encode(messages.GameStateResp{
			Error: &messages.Error{Message: err.Error()},
		})
	}

	toResume.Status = game.InProgress
	c.GameState = toResume

	return c.Encoder.Encode(messages.GameStateResp{State: *c.GameState})
}

// guessHandler updates the characters guessed or the characters missed by comparing
// the user guess against the secret word. The function ensures that if a characters
// appears multiple times in the secrete word the count will be update accordingly.
// If the user attempts to use a wrong charcater that had already been tried the character
// will not be counted as a miss.
func (c *controller) guessHandler(charGuessed string) error {
	c.System.Logger.Printf("%s is guessing %s", c.UserID, charGuessed)

	gameError := c.validateGameStatus()
	if gameError != nil {
		return c.Encoder.Encode(messages.GameStateResp{
			Error: gameError,
		})
	}

	for _, char := range charGuessed {
		charStr := string(char)

		if strings.Contains(c.GameState.WordToGuess, charStr) {
			occurences := strings.Count(c.GameState.WordToGuess, charStr)

			for occurences > 0 {
				c.GameState.CharsGuessed = append(c.GameState.CharsGuessed, charStr)
				occurences--
			}
		} else {
			if !utils.Contains(c.GameState.CharsTried, charStr) {
				c.GameState.CharsTried = append(c.GameState.CharsTried, charStr)
			}
		}
	}

	c.GameState.Status = c.updateGameStatus()

	return c.Encoder.Encode(messages.GameStateResp{
		State: *c.GameState,
	})
}

// validateGameStatus ensure the user can guess a character. Error messages are returned
// if a user tries to guess a hero but the game status doesn't allow it. This can happen if
// - the game hasn't started
// - there is no game in progress (e.g. all games have been paused or have been finished)
func (c controller) validateGameStatus() *messages.Error {
	if c.GameState == nil || c.GameState.Status != game.InProgress {
		return &messages.Error{Message: "you must start a new game or resume a paused game before guessing the hero"}
	}

	return nil
}

// updateGameStatus checks if the current game status should be set to game-over, win
// or in-progress.
func (c *controller) updateGameStatus() game.Status {
	if len(c.GameState.CharsTried) >= game.MaxWrongChars {
		return game.GameOver
	}

	if len(c.GameState.CharsGuessed) >= len(c.GameState.WordToGuess) {
		return game.Won
	}

	return game.InProgress
}

// handlePlayerAction calls the appropriate handle according to the command issued by
// the player. If no handler is found an error will be returned.
func (c *controller) handlePlayerAction(input messages.PlayerReq) error {

	switch input.Action {

	case game.Login:
		return c.loginHandler(input.Value)

	case game.NewGame:
		return c.newGameHandler()

	case game.Help:
		return c.helpHandler()

	case game.ListGames:
		return c.listGamesHandler()

	case game.ResumeGame:
		return c.resumeGameHandler(input.Value)

	case game.Guess:
		return c.guessHandler(input.Value)

	default:
		return fmt.Errorf("unexpected error: action %v was not recognized", input.Action)
	}

	return nil
}

// parseUserInput decodes and parses the incoming user request.
func parseUserInput(conn io.Reader) (*messages.PlayerReq, error) {
	var req messages.PlayerReq

	d := json.NewDecoder(conn)
	err := d.Decode(&req)
	if err != nil {
		return nil, err
	}

	req.Value = strings.ToLower(req.Value)

	return &req, nil
}
