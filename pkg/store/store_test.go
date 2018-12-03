package store

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Popcore/hangmango/pkg/game"
)

func TestSaveGame(t *testing.T) {
	store := memStore{
		games: make(map[string]map[int]game.State),
	}

	toSave := game.State{
		WordToGuess:  "test-word",
		CharsGuessed: []string{},
	}

	// save
	saved, err := store.SaveGame("user-id", toSave)
	expected := game.State{
		GameID:       1,
		WordToGuess:  "test-word",
		CharsGuessed: []string{},
	}

	assert.Nil(t, err)
	assert.Equal(t, store.games["user-id"], map[int]game.State{
		1: expected,
	})
	assert.Equal(t, expected, *saved)

	toSave2 := game.State{
		WordToGuess: "another-test-word",
	}
	saved2, err := store.SaveGame("user-id", toSave2)
	assert.Nil(t, err)
	assert.Equal(t, saved2.GameID, 2)
	assert.Len(t, store.games["user-id"], 2)

	// upsert
	toUpdate := game.State{
		GameID:       1,
		WordToGuess:  "test-word",
		CharsGuessed: []string{"a"},
	}

	_, err = store.SaveGame("user-id", toUpdate)
	assert.Nil(t, err)
	assert.Equal(t, store.games["user-id"][1], toUpdate)
	assert.Len(t, store.games["user-id"], 2)
}

func TestGetGameByID(t *testing.T) {
	store := NewMemStore()

	toSave := game.State{
		GameID:       1,
		WordToGuess:  "test-word",
		CharsGuessed: []string{},
	}

	_, err := store.SaveGame("user-id", toSave)
	assert.Nil(t, err)

	// wrong user
	_, err = store.GetGameByID("i-dont-exist", 99)
	assert.Equal(t, ErrorUserNotFound, err)

	// wrong id
	_, err = store.GetGameByID("user-id", 99)
	assert.Equal(t, ErrorGameNotFound, err)

	got, err := store.GetGameByID("user-id", 1)
	assert.Equal(t, toSave, *got)
}

func TestGetGameByUser(t *testing.T) {
	store := NewMemStore()

	toSave := game.State{
		GameID:       1,
		WordToGuess:  "test-word",
		CharsGuessed: []string{},
	}

	toSave2 := game.State{
		GameID:       2,
		WordToGuess:  "another-test-word",
		CharsGuessed: []string{},
	}

	_, err := store.SaveGame("user-id", toSave)
	assert.Nil(t, err)

	_, err = store.SaveGame("user-id", toSave2)
	assert.Nil(t, err)

	// wrong user
	_, err = store.GetGamesByUser("i-dont-exist")
	assert.Equal(t, ErrorUserNotFound, err)

	got, err := store.GetGamesByUser("user-id")
	assert.Len(t, got, 2)
}

func TestSaveNewUser(t *testing.T) {
	store := memStore{
		games: make(map[string]map[int]game.State),
	}

	err := store.SaveNewUser("user-id")
	assert.Nil(t, err)

	assert.Equal(t, store.games["user-id"], map[int]game.State{})
}
