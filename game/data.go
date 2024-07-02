package game

import (
	"time"

	"github.com/maria-mz/bash-battle-server/config"
)

type GameData struct {
	Config     config.GameConfig
	Challenges map[int]Challenge
	Players    map[string]*Player
}

func NewGameData(config config.GameConfig) *GameData {
	return &GameData{
		Config:     config,
		Challenges: GenerateChallenges(config),
		Players:    make(map[string]*Player),
	}
}

func (data *GameData) HasPlayer(name string) bool {
	_, ok := data.Players[name]
	return ok
}

func (data *GameData) GetPlayer(name string) (Player, bool) {
	player, ok := data.Players[name]
	return *player, ok
}

func (data *GameData) GetPlayers() []*Player {
	players := make([]*Player, len(data.Players))

	for _, player := range data.Players {
		players = append(players, player)
	}

	return players
}

func (data *GameData) NumPlayers() int {
	return len(data.Players)
}

func (data *GameData) AddPlayer(player *Player) {
	data.Players[player.Name] = player
}

func (data *GameData) RemovePlayer(name string) {
	delete(data.Players, name)
}

func (data *GameData) GetChallenge(round int) (Challenge, bool) {
	challenge, ok := data.Challenges[round]
	return challenge, ok
}

func (data *GameData) GetRoundDuration() time.Duration {
	return time.Duration(data.Config.RoundDuration) * time.Second
}

func (data *GameData) GetCountdownDuration() time.Duration {
	return time.Duration(data.Config.CountdownDuration) * time.Second
}

func (data *GameData) IsGameFull() bool {
	return len(data.Players) == data.Config.MaxPlayers
}

func (data *GameData) IsGameEmpty() bool {
	return len(data.Players) == 0
}

func (data *GameData) IsRoundValid(round int) bool {
	return round < 1 || round > data.Config.Rounds
}
