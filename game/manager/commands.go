package manager

type GameManagerCmd interface{}

type LoadRound struct {
	Round int
}

type SendRoundStartTime struct {
	Round int
}

type SendRoundEndTime struct {
	Round int
}

type SubmitScore struct {
	Round int
}
