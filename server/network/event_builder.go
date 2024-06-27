package network

import (
	"time"

	pb "github.com/maria-mz/bash-battle-proto/proto"
	"github.com/maria-mz/bash-battle-server/game"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func BuildPlayerJoinedEvent(player *game.Player) *pb.Event {
	event := &pb.Event{
		Event: &pb.Event_PlayerJoined{
			PlayerJoined: &pb.PlayerJoined{Player: player.ToProto()},
		},
	}

	return event
}

func BuildPlayerLeftEvent(player *game.Player) *pb.Event {
	event := &pb.Event{
		Event: &pb.Event_PlayerLeft{
			PlayerLeft: &pb.PlayerLeft{Player: player.ToProto()},
		},
	}

	return event
}

func BuildCountingDownEvent(round int, startsAt time.Time) *pb.Event {
	event := &pb.Event{
		Event: &pb.Event_CountingDown{
			CountingDown: &pb.CountingDown{
				RoundNumber: int32(round),
				StartsAt:    timestamppb.New(startsAt),
			},
		},
	}

	return event
}

func BuildRoundStartedEvent(round int, endsAt time.Time) *pb.Event {
	event := &pb.Event{
		Event: &pb.Event_RoundStarted{
			RoundStarted: &pb.RoundStarted{
				RoundNumber: int32(round),
				EndsAt:      timestamppb.New(endsAt),
			},
		},
	}

	return event
}

func BuildLoadRoundEvent(round int, challenge game.Challenge) *pb.Event {
	event := &pb.Event{
		Event: &pb.Event_LoadRound{
			LoadRound: &pb.LoadRound{
				RoundNumber: int32(round),
				Question:    challenge.Question,
				// TODO: Add files as bytes
			},
		},
	}

	return event
}

func BuildSubmitRoundScoreEvent() *pb.Event {
	event := &pb.Event{
		Event: &pb.Event_SubmitRoundScore{},
	}

	return event
}

func BuildGameOverEvent() *pb.Event {
	event := &pb.Event{
		Event: &pb.Event_GameOver{
			GameOver: &pb.GameOver{},
		},
	}

	return event
}
