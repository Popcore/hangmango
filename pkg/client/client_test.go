package client

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Popcore/hangmango/pkg/client/drawing"
	"github.com/Popcore/hangmango/pkg/game"
	"github.com/Popcore/hangmango/pkg/messages"
)

func TestNewGameRequest(t *testing.T) {

	// wConn and rConn are out mocked net.Conn and allow write and read operations
	// witouth greating a server instance.
	wConn, rConn := net.Pipe()
	defer wConn.Close()
	defer rConn.Close()

	resp := messages.GameStateResp{
		Error: nil,
		State: game.State{
			WordToGuess:  "guess",
			CharsGuessed: []string{},
			CharsTried:   []string{},
		},
	}

	go func() {
		json.NewEncoder(wConn).Encode(resp)
	}()

	var buf bytes.Buffer
	client := Client{
		Output:  &buf,
		Encoder: json.NewEncoder(ioutil.Discard),
		Decoder: json.NewDecoder(rConn),
	}

	client.handleUserCommands(messages.PlayerReq{Action: game.NewGame})

	assert.Contains(t, buf.String(), "Guess the hero: _ _ _ _ _")
	assert.Contains(t, buf.String(), "Characters tried")
}

func TestNewGameRequestError(t *testing.T) {
	wConn, rConn := net.Pipe()
	defer wConn.Close()
	defer rConn.Close()

	resp := messages.GameStateResp{
		Error: &messages.Error{
			Message: "the error message",
		},
	}

	go func() {
		json.NewEncoder(wConn).Encode(resp)
	}()

	var buf bytes.Buffer
	client := Client{
		Output:  &buf,
		Encoder: json.NewEncoder(ioutil.Discard),
		Decoder: json.NewDecoder(rConn),
	}

	client.handleUserCommands(messages.PlayerReq{Action: game.NewGame})

	assert.Contains(t, buf.String(), "the error message")
	assert.NotContains(t, buf.String(), "Guess the hero")
}

func TestHelpRequest(t *testing.T) {
	wConn, rConn := net.Pipe()
	defer wConn.Close()
	defer rConn.Close()

	resp := messages.HelpResp{
		Info: "the info message",
	}

	go func() {
		json.NewEncoder(wConn).Encode(resp)
	}()

	var buf bytes.Buffer
	client := Client{
		Output:  &buf,
		Encoder: json.NewEncoder(ioutil.Discard),
		Decoder: json.NewDecoder(rConn),
	}

	client.handleUserCommands(messages.PlayerReq{Action: game.Help})

	assert.Contains(t, buf.String(), "the info message")
}

func TestListGamesRequest(t *testing.T) {
	wConn, rConn := net.Pipe()
	defer wConn.Close()
	defer rConn.Close()

	resp := messages.ListGamesResp{
		Games: []game.State{
			{
				GameID:      1,
				WordToGuess: "foo",
				CharsTried:  []string{"a", "b"},
				Status:      game.Paused,
			},
			{
				GameID:      2,
				WordToGuess: "bar",
				CharsTried:  []string{"c", "d"},
				Status:      game.Won,
			},
		},
		Error: nil,
	}

	go func() {
		json.NewEncoder(wConn).Encode(resp)
	}()

	var buf bytes.Buffer
	client := Client{
		Output:  &buf,
		Encoder: json.NewEncoder(ioutil.Discard),
		Decoder: json.NewDecoder(rConn),
	}

	client.handleUserCommands(messages.PlayerReq{Action: game.ListGames})

	assert.Contains(t, buf.String(), "Game ID: 1 * Hero: _ _ _  * Characters tried: [a b] * Status: paused")
	assert.Contains(t, buf.String(), "Game ID: 2 * Hero: _ _ _  * Characters tried: [c d] * Status: won")
}

func TestListGamesRequestErr(t *testing.T) {
	wConn, rConn := net.Pipe()
	defer wConn.Close()
	defer rConn.Close()

	resp := messages.ListGamesResp{
		Games: []game.State{},
		Error: &messages.Error{
			Message: "the error message",
		},
	}

	go func() {
		json.NewEncoder(wConn).Encode(resp)
	}()

	var buf bytes.Buffer
	client := Client{
		Output:  &buf,
		Encoder: json.NewEncoder(ioutil.Discard),
		Decoder: json.NewDecoder(rConn),
	}

	client.handleUserCommands(messages.PlayerReq{Action: game.ListGames})

	assert.Contains(t, buf.String(), "the error message")
}

func TestResumeGamesRequest(t *testing.T) {
	wConn, rConn := net.Pipe()
	defer wConn.Close()
	defer rConn.Close()

	resp := messages.GameStateResp{
		State: game.State{
			GameID:      1,
			WordToGuess: "foo",
			CharsTried:  []string{"a"},
			Status:      game.InProgress,
		},
		Error: nil,
	}

	go func() {
		json.NewEncoder(wConn).Encode(resp)
	}()

	var buf bytes.Buffer
	client := Client{
		Output:  &buf,
		Encoder: json.NewEncoder(ioutil.Discard),
		Decoder: json.NewDecoder(rConn),
	}

	client.handleUserCommands(messages.PlayerReq{Action: game.ResumeGame, Value: "1"})

	assert.Contains(t, buf.String(), "Guess the hero: _ _ _ ")
	assert.Contains(t, buf.String(), drawing.Display[1])
	assert.Contains(t, buf.String(), "Characters tried: a")
}

func TestResumeGamesRequestErr(t *testing.T) {
	wConn, rConn := net.Pipe()
	defer wConn.Close()
	defer rConn.Close()

	resp := messages.ListGamesResp{
		Games: []game.State{},
		Error: &messages.Error{
			Message: "the error message",
		},
	}

	go func() {
		json.NewEncoder(wConn).Encode(resp)
	}()

	var buf bytes.Buffer
	client := Client{
		Output:  &buf,
		Encoder: json.NewEncoder(ioutil.Discard),
		Decoder: json.NewDecoder(rConn),
	}

	client.handleUserCommands(messages.PlayerReq{Action: game.ResumeGame, Value: "1"})

	assert.Contains(t, buf.String(), "the error message")
}

func TestGuessRequestRequest(t *testing.T) {
	wConn, rConn := net.Pipe()
	defer wConn.Close()
	defer rConn.Close()

	resp := messages.GameStateResp{
		State: game.State{
			GameID:      1,
			WordToGuess: "foo",
			CharsTried:  []string{"a", "b"},
			Status:      game.InProgress,
		},
		Error: nil,
	}

	go func() {
		json.NewEncoder(wConn).Encode(resp)
	}()

	var buf bytes.Buffer
	client := Client{
		Output:  &buf,
		Encoder: json.NewEncoder(ioutil.Discard),
		Decoder: json.NewDecoder(rConn),
	}

	client.handleUserCommands(messages.PlayerReq{Action: game.Guess, Value: "a"})

	assert.Contains(t, buf.String(), "Guess the hero: _ _ _ ")
	assert.Contains(t, buf.String(), drawing.Display[2])
	assert.Contains(t, buf.String(), "Characters tried: a - b")
}

func TestGuessRequestErr(t *testing.T) {
	wConn, rConn := net.Pipe()
	defer wConn.Close()
	defer rConn.Close()

	resp := messages.ListGamesResp{
		Games: []game.State{},
		Error: &messages.Error{
			Message: "the error message",
		},
	}

	go func() {
		json.NewEncoder(wConn).Encode(resp)
	}()

	var buf bytes.Buffer
	client := Client{
		Output:  &buf,
		Encoder: json.NewEncoder(ioutil.Discard),
		Decoder: json.NewDecoder(rConn),
	}

	client.handleUserCommands(messages.PlayerReq{Action: game.Guess, Value: "a"})

	assert.Contains(t, buf.String(), "the error message")
}
