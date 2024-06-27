package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBitmap(t *testing.T) {
	bitmap := NewActivityBitmap()
	assert.NotNil(t, bitmap)
}

func TestBitmapSetAck(t *testing.T) {
	bitmap := NewActivityBitmap()

	bitmap.SetLoadAck("client-1", true)
	bitmap.SetLoadAck("client-2", true)
	bitmap.SetSubmissionAck("client-1", true)

	assert.Equal(t, bitmap.CountAcks(AckLoad), 2)
	assert.Equal(t, bitmap.CountAcks(AckSubmission), 1)
}

func TestBitmapResetAcks(t *testing.T) {
	bitmap := NewActivityBitmap()

	bitmap.SetLoadAck("client-1", true)
	bitmap.SetSubmissionAck("client-1", true)
	bitmap.ResetAcks()

	assert.Equal(t, bitmap.CountAcks(AckLoad), 0)
	assert.Equal(t, bitmap.CountAcks(AckSubmission), 0)
}

func TestBitmapPop(t *testing.T) {
	bitmap := NewActivityBitmap()

	bitmap.SetLoadAck("client-1", true)
	bitmap.SetLoadAck("client-2", true)
	bitmap.SetSubmissionAck("client-1", true)
	bitmap.Pop("client-2")

	assert.Equal(t, bitmap.CountAcks(AckLoad), 1)
	assert.Equal(t, bitmap.CountAcks(AckSubmission), 1)
}
