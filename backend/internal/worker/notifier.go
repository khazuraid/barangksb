package worker

import (
	"context"
	"log/slog"
)

// TelegramNotifier matches the interface required by HTTP handlers
type TelegramNotifier interface {
	NotifyMovement(ctx context.Context, txType, sku, name string, qty int32, unit string, prev, next int32, person, notes string)
	SyncBotCommands(ctx context.Context) error
}

// QueuedTelegramNotifier sends notifications through Redis worker queue
type QueuedTelegramNotifier struct {
	underlying TelegramNotifier
	queue      *Queue
}

// NewQueuedTelegramNotifier wraps TelegramNotifier with Redis queue capability
func NewQueuedTelegramNotifier(underlying TelegramNotifier, queue *Queue) *QueuedTelegramNotifier {
	return &QueuedTelegramNotifier{
		underlying: underlying,
		queue:      queue,
	}
}

func (q *QueuedTelegramNotifier) NotifyMovement(ctx context.Context, txType, sku, name string, qty int32, unit string, prev, next int32, person, notes string) {
	if q.queue != nil && q.queue.rdb != nil {
		err := q.queue.Enqueue(ctx, TaskTypeTelegramMovement, TelegramMovementPayload{
			TxType:   txType,
			SKU:      sku,
			Name:     name,
			Quantity: qty,
			Unit:     unit,
			Prev:     prev,
			Next:     next,
			Person:   person,
			Notes:    notes,
		})
		if err == nil {
			return
		}
		slog.Warn("enqueue telegram movement failed, falling back to direct", "err", err)
	}

	if q.underlying != nil {
		go q.underlying.NotifyMovement(ctx, txType, sku, name, qty, unit, prev, next, person, notes)
	}
}

func (q *QueuedTelegramNotifier) SyncBotCommands(ctx context.Context) error {
	if q.underlying != nil {
		return q.underlying.SyncBotCommands(ctx)
	}
	return nil
}
