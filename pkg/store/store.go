package store

import (
	"errors"
	"sync"

	"github.com/Popcore/hangmango/pkg/game"
)

var (
	ErrorGameNotFound  = errors.New("game not found")
	ErrorUserNotFound  = errors.New("user not found")
	ErrorMissingGameID = errors.New("ensure game has a valid ID. 0 is not a valid value")
)

// Storer defines the functionalies a data store must expose in order to
// allow games to be saved, queried, deleted etc. Its main funtionalies focus
// on keeping track of the active games and players.
type Storer interface {
	SaveGame(userID string, g game.State) (*game.State, error)

	SaveNewUser(userID string) error

	GetGameByID(userID string, gameID int) (*game.State, error)

	GetGamesByUser(userID string) ([]game.State, error)
}

// memStore is the in-memory implementation of the Storer interface. The embedded
// mutex ensures protected concurrent access to its underlying games map.
type memStore struct {
	sync.Mutex
	games map[string]map[int]game.State
}

// NewMemStore instatiate a new memory store. The games map expects a user-id
// as key an a collection of games as values.
func NewMemStore() Storer {
	return &memStore{
		games: make(map[string]map[int]game.State),
	}
}

// SaveGame saves a new game or upserts an existing one. Internally SaveGame checks
// if g contains a valid id, and if if doesn't a new game will be saved with a new
// id assigned to it. If g contains the id the existing game will be updated.
func (s *memStore) SaveGame(userID string, g game.State) (*game.State, error) {
	s.Lock()
	defer s.Unlock()

	if g.GameID == 0 {
		g.GameID = len(s.games[userID]) + 1
	}

	// new user
	games, ok := s.games[userID]
	if !ok {
		s.games[userID] = map[int]game.State{
			g.GameID: g,
		}

		return &g, nil
	}

	games[g.GameID] = g

	return &g, nil
}

// SaveNewUser adds a new user as a key to the memory store and initialize its
// map of games-id - games info
func (s *memStore) SaveNewUser(userID string) error {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.games[userID]; !ok {
		s.games[userID] = make(map[int]game.State)
	}

	return nil
}

// GetGameByID returns a game identified by its id and owned by userID. Returns an error
// if the user or the game cannot be found.
func (s *memStore) GetGameByID(userID string, gameID int) (*game.State, error) {
	s.Lock()
	defer s.Unlock()

	games, ok := s.games[userID]
	if !ok {
		return nil, ErrorUserNotFound
	}

	for id, game := range games {
		if id == gameID {
			return &game, nil
		}
	}

	return nil, ErrorGameNotFound
}

// GetGamesByUser returns a list of games owned by userID. Returns an error if the
// user cannot be found.
func (s *memStore) GetGamesByUser(userID string) ([]game.State, error) {
	s.Lock()
	defer s.Unlock()

	var gameSlice = []game.State{}
	games, ok := s.games[userID]
	if !ok {
		return nil, ErrorUserNotFound
	}

	// flatten the map
	for _, game := range games {
		gameSlice = append(gameSlice, game)
	}

	return gameSlice, nil
}
