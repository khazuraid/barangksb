package telegram

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBotSenderNoTokenSafeNoOp(t *testing.T) {
	b := NewBot("", "12345,bad,0", nil)
	assert.True(t, b.chatIDs[12345])
	assert.Len(t, b.chatIDs, 1)
	b.InitSender(context.Background())
	assert.Nil(t, b.botInst)
	// Must not panic when called without bot token or pool
	b.NotifyMovement(context.Background(), "ADJUST", "SKU-1", "Laptop", 2, "unit", 5, 7, "Admin", "Opname")
}
