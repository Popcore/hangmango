package handlers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Popcore/hangmango/pkg/game"
	"github.com/Popcore/hangmango/pkg/messages"
	"github.com/Popcore/hangmango/pkg/store"
)

func TestLoginHandler(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{})

	c := controller{
		System: System{
			Logger: log.New(ioutil.Discard, "event: ", log.LstdFlags),
			Store:  store.NewMemStore(),
		},
		Encoder: json.NewEncoder(buffer),
	}

	err := c.loginHandler("user-id")
	assert.Nil(t, err)

	resp := messages.HelpResp{}
	err = json.NewDecoder(buffer).Decode(&resp)
	assert.Nil(t, err)

	assert.Equal(t, game.Rules, resp.Info)
}

func TestNewGameHandler(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{})

	c := controller{
		System: System{
			Logger: log.New(ioutil.Discard, "event: ", log.LstdFlags),
			Store:  store.NewMemStore(),
		},
		Encoder: json.NewEncoder(buffer),
	}

	c.System.Store.SaveNewUser("user-id")
	c.UserID = "user-id"

	err := c.newGameHandler()
	assert.Nil(t, err)

	var resp messages.GameStateResp

	err = json.NewDecoder(buffer).Decode(&resp)
	assert.Nil(t, err)

	assert.Equal(t, game.InProgress, resp.State.Status)
	assert.Equal(t, []string{}, resp.State.CharsTried)
	assert.Contains(t, resp.State.WordToGuess, "_")

	err = c.newGameHandler()
	assert.Nil(t, err)

	games, err := c.System.Store.GetGamesByUser("user-id")
	assert.Nil(t, err)

	assert.Len(t, games, 2)

}

func TestHelpHandler(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{})

	c := controller{
		System: System{
			Logger: log.New(ioutil.Discard, "event: ", log.LstdFlags),
			Store:  store.NewMemStore(),
		},
		Encoder: json.NewEncoder(buffer),
	}

	c.System.Store.SaveNewUser("user-id")
	c.UserID = "user-id"

	err := c.helpHandler()
	assert.Nil(t, err)

	var resp messages.HelpResp

	err = json.NewDecoder(buffer).Decode(&resp)
	assert.Nil(t, err)

	assert.Equal(t, game.Rules, resp.Info)
}

func TestResumeGameHandler(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{})

	c := controller{
		System: System{
			Logger: log.New(ioutil.Discard, "event: ", log.LstdFlags),
			Store:  store.NewMemStore(),
		},
		Encoder: json.NewEncoder(buffer),
	}

	c.System.Store.SaveNewUser("user-id")
	c.UserID = "user-id"

	g, err := c.System.Store.SaveGame("user-id", game.State{
		WordToGuess: "foo",
		Status:      game.InProgress,
	})
	assert.Nil(t, err)

	err = c.resumeGameHandler(strconv.Itoa(g.GameID))
	assert.Nil(t, err)

	var resp messages.GameStateResp

	err = json.NewDecoder(buffer).Decode(&resp)
	assert.Nil(t, err)

	expected := g
	expected.WordToGuess = "_ _ _ "
	expected.Status = game.InProgress

	assert.Equal(t, *expected, resp.State)
	assert.Equal(t, c.GameState.GameID, g.GameID)
}

func TestGuessHandlerGameOver(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{})

	c := controller{
		System: System{
			Logger: log.New(ioutil.Discard, "event: ", log.LstdFlags),
			Store:  store.NewMemStore(),
		},
		Encoder: json.NewEncoder(buffer),
	}

	c.System.Store.SaveNewUser("user-id")
	c.UserID = "user-id"

	c.GameState = &game.State{
		WordToGuess:  "foo",
		Status:       game.InProgress,
		CharsGuessed: []string{"f"},
		CharsTried:   []string{"a", "b", "c", "d", "e"},
	}

	err := c.guessHandler("g")
	assert.Nil(t, err)

	var resp messages.GameStateResp

	err = json.NewDecoder(buffer).Decode(&resp)
	assert.Nil(t, err)
	assert.Equal(t, game.InProgress, resp.State.Status)

	err = c.guessHandler("h")
	assert.Nil(t, err)

	err = json.NewDecoder(buffer).Decode(&resp)
	assert.Nil(t, err)
	assert.Equal(t, game.GameOver, resp.State.Status)
}

func TestGuessHandlerGameWon(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{})

	c := controller{
		System: System{
			Logger: log.New(ioutil.Discard, "event: ", log.LstdFlags),
			Store:  store.NewMemStore(),
		},
		Encoder: json.NewEncoder(buffer),
	}

	c.System.Store.SaveNewUser("user-id")
	c.UserID = "user-id"

	c.GameState = &game.State{
		WordToGuess:  "foo",
		Status:       game.InProgress,
		CharsGuessed: []string{"f", "o"},
		CharsTried:   []string{"a", "b", "c", "d", "e"},
	}

	err := c.guessHandler("o")
	assert.Nil(t, err)

	var resp messages.GameStateResp

	err = json.NewDecoder(buffer).Decode(&resp)
	assert.Nil(t, err)
	assert.Equal(t, game.Won, resp.State.Status)
}

func TestValidateGameStatus(t *testing.T) {
	testcases := []struct {
		state    *game.State
		expected *messages.Error
	}{
		{
			state:    nil,
			expected: &messages.Error{Message: "you must start a new game or resume a paused game before guessing the hero"},
		},
		{
			state: &game.State{
				Status: game.Paused,
			},
			expected: &messages.Error{Message: "you must start a new game or resume a paused game before guessing the hero"},
		},
		{
			state: &game.State{
				Status: game.GameOver,
			},
			expected: &messages.Error{Message: "you must start a new game or resume a paused game before guessing the hero"},
		},
	}

	buffer := bytes.NewBuffer([]byte{})

	c := controller{
		System: System{
			Logger: log.New(ioutil.Discard, "event: ", log.LstdFlags),
			Store:  store.NewMemStore(),
		},
		Encoder: json.NewEncoder(buffer),
	}

	c.System.Store.SaveNewUser("user-id")
	c.UserID = "user-id"

	for _, testcase := range testcases {
		c.GameState = testcase.state

		err := c.guessHandler("o")
		assert.Nil(t, err)

		var resp messages.GameStateResp
		err = json.NewDecoder(buffer).Decode(&resp)
		assert.Nil(t, err)
		assert.Equal(t, testcase.expected, resp.Error)
	}
}

func TestParseUserInput(t *testing.T) {
	testcases := []struct {
		input    []byte
		expected *messages.PlayerReq
	}{
		{
			input: []byte(`{"action": "try", "value": "x"}`),
			expected: &messages.PlayerReq{
				Action: "try",
				Value:  "x",
			},
		},
		{
			input: []byte(`{"action": "new"}`),
			expected: &messages.PlayerReq{
				Action: "new",
				Value:  "",
			},
		},
	}

	for _, testcase := range testcases {
		reader := bytes.NewReader(testcase.input)
		got, err := parseUserInput(reader)

		assert.Nil(t, err)
		assert.Equal(t, testcase.expected, got)
	}
}
