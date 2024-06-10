package main

import (
	"errors"

	"github.com/gorilla/websocket"
)

type ChessGame struct {
	whiteWebsocket *websocket.Conn
	blackWebsocket *websocket.Conn
}

type Message struct {
	Type      string `json:"type" validate:"required,oneof=start move error"`
	Color     string `json:"color" validate:"oneof=white black,required_if=Type start"`
	From      string `json:"from" validate:"required_if=Type move"`
	To        string `json:"to" validate:"required_if=Type move"`
	Promotion string `json:"promotion" validate:"oneof=q r b k,required_if=Type move"`
}

func NewChessGame(ws *websocket.Conn) *ChessGame {
	game := ChessGame{whiteWebsocket: ws}
	return &game
}

var ErrCannotJoinStartedGame = errors.New("cannot join a started game")

func (game *ChessGame) Join(ws *websocket.Conn) error {
	// you cannot join the same game twice
	if game.blackWebsocket != nil {
		return ErrCannotJoinStartedGame
	}
	game.blackWebsocket = ws
	whiteChannel := make(chan Message)
	blackChannel := make(chan Message)
	go playChess(game.whiteWebsocket, game.blackWebsocket, whiteChannel, blackChannel)
	go forwardFromWebsocketToChannel(game.whiteWebsocket, whiteChannel)
	go forwardFromWebsocketToChannel(game.blackWebsocket, blackChannel)
	return nil
}

func playChess(
	whiteWebsocket, blackWebsocket *websocket.Conn,
	whiteChannel, blackChannel <-chan Message,
) {
	turnWhite := true
	whiteWebsocket.WriteJSON(Message{Type: "start", Color: "white"})
	blackWebsocket.WriteJSON(Message{Type: "start", Color: "black"})
	for {
		select {
		case message := <-whiteChannel:
			if message.Type == "error" {
				return
			}
			if turnWhite {
				blackWebsocket.WriteJSON(message)
				turnWhite = false
			}
		case message := <-blackChannel:
			if message.Type == "error" {
				return
			}
			if !turnWhite {
				whiteWebsocket.WriteJSON(message)
				turnWhite = true
			}
		}
	}
}

func forwardFromWebsocketToChannel(ws *websocket.Conn, ch chan<- Message) {
	defer ws.Close()
	for {
		message := Message{}
		err := ws.ReadJSON(&message)

		if err != nil {
			ch <- Message{Type: "error"}
			return
		}

		ch <- message
	}
}
