package state

import (
	"reflect"
	"testing"
)

var infoOne = RoundInfo{
	Question:   "question 1",
	InputFile:  "/path/to/input1.txt",
	OutputFile: "/path/to/output1.txt",
}

var infoTwo = RoundInfo{
	Question:   "question 2",
	InputFile:  "/path/to/input2.txt",
	OutputFile: "/path/to/output2.txt",
}

func initGamePlan() GamePlan {
	plan := NewGamePlan()
	plan.AddRound(infoOne)
	plan.AddRound(infoTwo)
	return plan
}

func initGameConfig() GameConfig {
	return GameConfig{
		Plan:         initGamePlan(),
		RoundSeconds: 300,
	}
}

func initGameStore() GameStore {
	store := *NewGameStore(initGameConfig())
	store.AddPlayer("player-0")
	return store
}

func TestAddPlayer(t *testing.T) {
	store := initGameStore()

	// Case 1: Name is new
	err := store.AddPlayer("player-1")

	if err != nil {
		t.Fatalf("expected no error, but got %s", err)
	}

	// Case 2: Name is in use
	err = store.AddPlayer("player-1")

	if err == nil {
		t.Fatalf("expected error")
	}

	t.Log("err = ", err)
}

func TestGetRoundInfo(t *testing.T) {
	store := initGameStore()

	// Case 1: Round exists
	num := RoundNumber(1)
	info, ok := store.GetRoundInfo(num)

	if !ok {
		t.Fatalf("expected ok, but got not ok")
	}

	if !reflect.DeepEqual(info, infoOne) {
		t.Fatalf("infos don't match")
	}

	// Case 2: Round doesn't exist
	num = RoundNumber(200)
	_, ok = store.GetRoundInfo(num)

	if ok {
		t.Fatalf("expected not ok, but got ok")
	}
}

func TestSetPlayerRoundStat(t *testing.T) {
	store := initGameStore()

	stat1 := RoundStat{
		WasBeat: true,
		Command: "awk '{print $1, $3}' input1.txt",
	}

	err := store.SetPlayerRoundStat("player-0", 1, stat1)

	if err != nil {
		t.Fatalf("expected no error, but got %s", err)
	}

	// Check if it was properly set
	stat2, ok := store.GetPlayerRoundStat("player-0", 1)

	if !ok {
		t.Fatalf("expected ok, but got not ok")
	}

	if !reflect.DeepEqual(stat1, stat2) {
		t.Fatalf("stats don't match")
	}
}
