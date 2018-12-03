package handlers

import (
	"math/rand"
	"time"
)

// heros is a collection of cartoon heroes that will be used as words
// to guess during a hangman game.
var (
	heroes = []string{
		"superman",
		"spiderman",
		"batman",
		"catwoman",
		"jocker",
		"wolverine",
		"mickeymouse",
		"donaldduck",
		"wonderwoman",
		"capitanplanet",
		"rickandmorty",
		"ericcartman",
	}
)

// randomHero returns a random character from the heroes list.
func randomHero() string {
	rand.Seed(time.Now().Unix())
	n := rand.Int() % len(heroes)

	return heroes[n]
}
