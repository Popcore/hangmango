package client

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/Popcore/hangmango/pkg/client/drawing"
	"github.com/Popcore/hangmango/pkg/game"
	"github.com/Popcore/hangmango/pkg/messages"
)

// Client is responsible for connecting to the upstream server, transmitting
// the player actions and managing the server responses.
type Client struct {
	Port    string
	Output  io.Writer
	Encoder *json.Encoder
	Decoder *json.Decoder
}

// New returns a new client connected to the server and ready to play.
func New(port string) (*Client, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return nil, err
	}

	fmt.Println("Welcome to HangmanGo")

	return &Client{
		Port:    port,
		Output:  os.Stdout,
		Encoder: json.NewEncoder(conn),
		Decoder: json.NewDecoder(conn),
	}, nil
}

// Play starts a new gaming sessions. It authenticates the player and starts listening
// to the commnads issueed by her/him.
func (c Client) Play() error {
	err := c.authenticateUser()
	if err != nil {
		fmt.Fprintf(c.Output, "unexpected error authenticating user: %v", err)
		os.Exit(1)
	}

	return c.handleGameIO()
}

// parseUserCommand parses the player's input and returns a PlayerReq message type that
// can be encoded and sent to the server.
func (c Client) parseUserCommand() messages.PlayerReq {
	var req messages.PlayerReq

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Fprint(c.Output, "=> ")
		text, _ := reader.ReadString('\n')

		text = strings.TrimSpace(text)
		if text == "" {
			fmt.Fprintln(c.Output, "Please enter a valid command")
		}

		parts := strings.Split(text, " ")

		req = messages.PlayerReq{
			Action: game.PlayerAction(parts[0]),
		}

		if len(parts) > 1 {
			joined := strings.Join(parts[1:len(parts)], "")
			req.Value = strings.TrimRight(joined, "\r\n")
		}

		break
	}

	return req
}

// encodeRequest encode PlayerReq messages as JSON payloads. Returns an error if the
// encoding process fails.
func (c Client) encodeRequest(req messages.PlayerReq) error {
	return c.Encoder.Encode(req)
}

// decodeResponse decodes server JSON responses into resp. Returns an error if the
// decoding process fails.
func (c Client) decodeResponse(resp interface{}) error {
	return c.Decoder.Decode(&resp)
}

// getUserName prompts players to enter their user name. It returns the user name.
// Empty values are not valid user names.
func (c Client) getUserName() string {
	var username string
	for {
		fmt.Fprint(c.Output, "=> Enter your username: ")
		reader := bufio.NewReader(os.Stdin)

		u, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(c.Output, "Unexpected error: %v.\n", err)
		}

		u = strings.TrimSpace(strings.TrimRight(u, "\r\n"))
		if u != "" {
			username = u
			break
		}

		fmt.Fprintln(c.Output, "Please enter a valid user name")
	}

	return username
}

// authenticateUser sends a login request including the username to the upstream server.
func (c Client) authenticateUser() error {
	username := c.getUserName()

	err := c.encodeRequest(messages.PlayerReq{
		Action: game.Login,
		Value:  username,
	})
	if err != nil {
		return err
	}

	var resp messages.HelpResp
	err = c.decodeResponse(&resp)
	if err != nil {
		return err
	}

	if resp.Error != nil {
		return errors.New(resp.Error.Message)
	}
	fmt.Fprintln(c.Output, resp.Info)

	return nil
}

// newGameRequest sends a new game request to the server and displays the response.
func (c Client) newGameRequest() {
	err := c.Encoder.Encode(messages.PlayerReq{Action: game.NewGame})
	if err != nil {
		fmt.Fprintf(c.Output, "unexpected error: %v \n", err)
	}

	var resp messages.GameStateResp
	err = c.decodeResponse(&resp)
	if err != nil {
		fmt.Fprintf(c.Output, "unexpected error: %s \n", err)
	}

	if resp.Error != nil {
		fmt.Fprintln(c.Output, resp.Error)
	} else {
		fmt.Fprintf(c.Output, "Guess the hero: %s \n", resp.State.WordToGuess)
		fmt.Fprintln(c.Output, drawing.Display[len(resp.State.CharsTried)])
		fmt.Fprintf(c.Output, "Characters tried: %s \n", strings.Join(resp.State.CharsTried, " - "))
	}
}

// helpRequest sends a help request to the server and displays the response.
func (c Client) helpRequest() {
	err := c.encodeRequest(messages.PlayerReq{Action: game.Help})
	if err != nil {
		fmt.Fprintf(c.Output, "unexpected error: %v \n", err)
	}

	var resp messages.HelpResp
	err = c.decodeResponse(&resp)
	if err != nil {
		fmt.Printf("Unexpected error: %v", err)
	}

	fmt.Fprintln(c.Output, resp.Info)
}

// listGamesRequest sends a list games request to the server and displays the response.
func (c Client) listGamesRequest() {
	err := c.encodeRequest(messages.PlayerReq{Action: game.ListGames})
	if err != nil {
		fmt.Fprintf(c.Output, "unexpected error: %v \n", err)
	}

	var resp messages.ListGamesResp
	err = c.decodeResponse(&resp)
	if err != nil {
		fmt.Fprintf(c.Output, "unexpected error: %v", err)
	}

	if resp.Error != nil {
		fmt.Fprintln(c.Output, resp.Error.Message)
		return
	}

	if len(resp.Games) == 0 {
		fmt.Fprintf(c.Output, "no games have been found. Type '%v' to start \n", game.NewGame)
	} else {
		for _, g := range resp.Games {
			fmt.Fprintf(c.Output, "Game ID: %d * Hero: %s * Characters tried: %v * Status: %v \n", g.GameID, g.WordToGuess, g.CharsTried, g.Status)
		}
	}
}

// resumeGameRequest sends a resume games request to the server and displays the response.
// The request must contain the id of the game to resume.
func (c Client) resumeGameRequest(gameID string) {
	err := c.encodeRequest(messages.PlayerReq{Action: game.ResumeGame, Value: gameID})
	if err != nil {
		fmt.Fprintf(c.Output, "unexpected error: %v \n", err)
	}

	var resp messages.GameStateResp
	err = c.decodeResponse(&resp)
	if err != nil {
		fmt.Fprintf(c.Output, "Unexpected error: %v", err)
	}

	if resp.Error != nil {
		fmt.Fprintln(c.Output, resp.Error.Message)
		return
	}

	fmt.Fprintf(c.Output, "Guess the hero: %s \n", resp.State.WordToGuess)
	fmt.Fprintln(c.Output, drawing.Display[len(resp.State.CharsTried)])
	fmt.Fprintf(c.Output, "Characters tried: %s \n", strings.Join(resp.State.CharsTried, " - "))
}

// guessRequest sends a guess request to the server and displays the response.
// The request must contain the value to try against the hidden word.
func (c Client) guessRequest(guess string) {
	err := c.encodeRequest(messages.PlayerReq{Action: game.Guess, Value: guess})
	if err != nil {
		fmt.Fprintf(c.Output, "unexpected error: %v \n", err)
	}
	var resp messages.GameStateResp
	err = c.decodeResponse(&resp)
	if err != nil {
		fmt.Fprintf(c.Output, "Unexpected error: %v", err)
	}

	if resp.Error != nil {
		fmt.Fprintf(c.Output, "Error: %s \n", resp.Error.Message)
	} else {
		fmt.Fprintf(c.Output, "Guess the hero: %s \n", resp.State.WordToGuess)
		fmt.Fprintln(c.Output, drawing.Display[len(resp.State.CharsTried)])
		fmt.Fprintf(c.Output, "Characters tried: %s \n", strings.Join(resp.State.CharsTried, " - "))

		if resp.State.Status == game.GameOver {
			fmt.Fprintln(c.Output, "*** GAME OVER ***")
		}

		if resp.State.Status == game.Won {
			fmt.Fprintln(c.Output, "*** YOU WIN ***")
		}
	}
}

// handleUserCommands takes the command issued by the player inteh form of a request
// message and calls the approprioate action to perform according to the command type.
func (c Client) handleUserCommands(req messages.PlayerReq) {
	switch req.Action {
	case game.NewGame:
		c.newGameRequest()

	case game.Help:
		c.helpRequest()

	case game.ListGames:
		c.listGamesRequest()

	case game.ResumeGame:
		c.resumeGameRequest(req.Value)

	case game.Guess:
		c.guessRequest(req.Value)

	default:
		fmt.Fprintf(c.Output, "Unknown command. Type '%v' to see the available actions \n", game.Help)
	}
}

// handleGameIO listens to the players input commands and sends them off for porcessing
// to the upstream server.
func (c Client) handleGameIO() error {
	for {
		req := c.parseUserCommand()
		c.handleUserCommands(req)
	}
}
