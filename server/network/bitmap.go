package network

type ClientActivity int

const (
	AckLoad ClientActivity = iota
	AckSubmission
)

type ActivityBitmap map[string][]bool

func NewActivityBitmap() ActivityBitmap {
	return make(map[string][]bool)
}

func (bitmap ActivityBitmap) SetLoadAck(clientID string, loaded bool) {
	if _, ok := bitmap[clientID]; !ok {
		bitmap[clientID] = make([]bool, 2)
	}

	bitmap[clientID][AckLoad] = loaded
}

func (bitmap ActivityBitmap) SetSubmissionAck(clientID string, submitted bool) {
	if _, ok := bitmap[clientID]; !ok {
		bitmap[clientID] = make([]bool, 2)
	}

	bitmap[clientID][AckSubmission] = submitted
}

func (bitmap ActivityBitmap) ResetAcks() {
	for _, row := range bitmap {
		row[AckLoad] = false
		row[AckSubmission] = false
	}
}

func (bitmap ActivityBitmap) Pop(clientID string) {
	delete(bitmap, clientID)
}

func (bitmap ActivityBitmap) CountAcks(ack ClientActivity) int {
	count := 0

	for _, row := range bitmap {
		if row[ack] {
			count++
		}
	}

	return count
}

func (bitmap ActivityBitmap) GetStatus(ack ClientActivity, clientID string) bool {
	return bitmap[clientID][ack]
}
