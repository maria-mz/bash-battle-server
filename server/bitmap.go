package server

type ClientAck int

const (
	AckLoad ClientAck = iota
	AckSubmission
)

type ClientAckStatusBitmap map[string][]bool

func NewClientAckBitmap() ClientAckStatusBitmap {
	return make(map[string][]bool)
}

func (bitmap ClientAckStatusBitmap) SetLoadAck(clientID string, loaded bool) {
	if _, ok := bitmap[clientID]; !ok {
		bitmap[clientID] = make([]bool, 2)
	}

	bitmap[clientID][AckLoad] = loaded
}

func (bitmap ClientAckStatusBitmap) SetSubmissionAck(clientID string, submitted bool) {
	if _, ok := bitmap[clientID]; !ok {
		bitmap[clientID] = make([]bool, 2)
	}

	bitmap[clientID][AckSubmission] = submitted
}

func (bitmap ClientAckStatusBitmap) ResetAcks() {
	for _, row := range bitmap {
		row[AckLoad] = false
		row[AckSubmission] = false
	}
}

func (bitmap ClientAckStatusBitmap) Pop(clientID string) {
	delete(bitmap, clientID)
}

func (bitmap ClientAckStatusBitmap) CountAcks(ack ClientAck) int {
	count := 0

	for _, row := range bitmap {
		if row[ack] {
			count++
		}
	}

	return count
}
