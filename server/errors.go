package server

import "errors"

var ErrTokenNotRecognized = errors.New("token not recognized")
var ErrStreamAlreadyActive = errors.New("stream is already active")
var ErrTooManyPlayers = errors.New("max players already reached")
var ErrClientNotFound = errors.New("client not found in game")
var ErrUsernameTaken = errors.New("a player with this name already exists")
var ErrJoinOnGameStarted = errors.New("cannot join game: game already started")
var ErrStreamOnGameOver = errors.New("cannot stream game: game is over")
