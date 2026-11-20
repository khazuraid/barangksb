package worker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"inventariskantor/internal/config"
)

type mockTgNotifier struct {
	done   chan struct{}
	called bool
	txType string
	sku    string
}

func (m *mockTgNotifier) NotifyMovement(ctx context.Context, txType, sku, name string, qty int32, unit string, prev, next int32, person, notes string) {
	m.called = true
	m.txType = txType
	m.sku = sku
	if m.done != nil {
		close(m.done)
	}
}

func (m *mockTgNotifier) SyncBotCommands(ctx context.Context) error {
	return nil
}

func TestQueuedTelegramNotifier_FallbackWhenNoRedis(t *testing.T) {
	mock := &mockTgNotifier{done: make(chan struct{})}
	q := NewQueue(nil) // No redis
	notifier := NewQueuedTelegramNotifier(mock, q)

	notifier.NotifyMovement(context.Background(), "IN", "SKU-001", "Kertas", 5, "Rim", 10, 15, "Budi", "Pengadaan")

	select {
	case <-mock.done:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for fallback NotifyMovement")
	}

	assert.True(t, mock.called)
	assert.Equal(t, "IN", mock.txType)
	assert.Equal(t, "SKU-001", mock.sku)
}

func TestQueue_EnqueueNoRedis(t *testing.T) {
	q := NewQueue(nil)
	err := q.Enqueue(context.Background(), "test", map[string]string{"foo": "bar"})
	assert.NoError(t, err)
}

func TestNewRedisClient_URLValidation(t *testing.T) {
	// Invalid URL should return error
	cfg := &config.Config{
		RedisURL: "://invalid-url",
	}
	client, err := NewRedisClient(cfg)
	assert.Error(t, err)
	assert.Nil(t, client)
}
