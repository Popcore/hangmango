# HangManGo
HangManGo ia an implemenatation of the hangman game in Go.

## Setup

### Get the package
The easiest way to use HangManGo is via
```
go get github.com/Popcore/hangmango
```
which will clone the package on your GOPATH and install the hangmango executable.

The package comes with a Makefile that provides convenient shortcuts for building the package manually and running tests.
See `make help` for a list of the available make commands.

### Quick overview
Once installed (or built) the `hangmango` executable provieds a simple CLI the exposes two commands:
- `hangmango server` can be used to start a new server. In the context of the hangman game the server will provide words to guess and update the game status according to the player's input.
- `hangmango client` can be used to start a new client. The client allows players to interact with the server, play new games, resume existing ones and list the rules and the availbale commands.
By defualt the server will be available at port 9090 and the client will try to connect to the same port. This behaviour can be changed using the `-p` flag available for both server and client.
See `hangmango help` for a full description of the two commands and their options.

## Play the game
The game is played via the command line and the basic set-up requires two command line windows, one for the server and one for the client. Multiple clients can be started using more command line windows.

Before starting a new game ensure the server is up and running. If so clients (aka players) can start connnecting to it.

Once a client is connected it will be required to enter a username, after which the gaming session can finally start.
The first screen will display the rules of the game and the availabe commands the player can use to interact with the server.

## Design notes

### Components
The main components or services the constitute the game are

- a Server that accepts and handle players connections. Connections are handled in dedicated goroutines and once a new connection starts the server will continuosly listen for incoming messages. The server and its handlers are responsible for maintaing the state of games in progress, interact with the data store to save or retrieve games and log actions and errors for internal inspection.

- a Client that reads from stdin, builds messages to send to the server and parses and displays the responses it receives.
Beside parsing and cleaning of users' inputs the Client responsibilities include converting server responses into visual outputs to display back to players.

- a Data Store, used to save information about a player's games history. For simplicity the current implementation is a simple in-memory data store, but anything that satisfy the `Storer` interface (defined in pkg/store/store.go) can be used to replace it without affecting the other two componens.
The in-memory data store is implemented as a key-value map that uses player's user names as keys and a second map of games as values. In this seconf map games ids are the keys and game state their values.

### Networking protocol
Network communications happen over TCP. The protocol was chosen because of its ability to handle multiple connections between the server and its clients and beacuse it allows both endpoints to send and receive streams of bytes.

### Messaging protocol
The messaging protocol used to exchange messages between the server and the client is JSON. Messages sent by clients must contain a command the server can understand (e.g. start a new game, display help) and optional values (e.g. try character 'x').
Messages are defined in pkg/messages.

### Players authetication
In the current model authentication simply relies on the user name supplied by the player before starting a new session. No password or other forms of security checks involved. This solution is far from ideal as multiple players could connect concurrently to the server and authenticate using the same user name, but for simplicity sake we assume this is ok for now.

## Tests
Tests can be run via `make test`.
Coverage can be improved but the main input/output functionalities and outcomes produced by a player command have been tested on both server and client packages. Tests include error handling, especially on the client side. Read and write operations against the data store have also been tested together with some (basic) integration between the server and the store.

