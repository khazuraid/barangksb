package scheduler

import (
	"context"
	"log/slog"
	"time"
)

// Daily runs fn every day at hour:minute WIB.
func Daily(hour, minute int, fn func(ctx context.Context)) {
	go func() {
		ctx := context.Background()
		loc := time.FixedZone("WIB", 7*3600)
		for {
			now := time.Now().In(loc)
			next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, loc)
			if next.Before(now) {
				next = next.Add(24 * time.Hour)
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(next.Sub(now)):
			}
			fn(context.Background())
		}
	}()
}

// Hourly runs fn every hour.
func Hourly(ctx context.Context, fn func(ctx context.Context)) {
	go func() {
		t := time.NewTicker(time.Hour)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				fn(ctx)
			}
		}
	}()
}

// Monthly runs fn on day-of-month at 06:00 WIB.
func Monthly(day int, fn func(ctx context.Context)) {
	go func() {
		loc := time.FixedZone("WIB", 7*3600)
		for {
			now := time.Now().In(loc)
			next := time.Date(now.Year(), now.Month(), day, 6, 0, 0, 0, loc)
			if next.Before(now) {
				next = next.AddDate(0, 1, 0)
			}
			slog.Info("scheduler: next monthly run", "at", next.Format(time.RFC3339))
			time.Sleep(time.Until(next))
			fn(context.Background())
		}
	}()
}
