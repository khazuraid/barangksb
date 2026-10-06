package handlers

import (
	"sync"

	"golang.org/x/time/rate"
)

var (
	testLimOnce sync.Once
	testLimInst *rate.Limiter
)

func testLimiterAllow() bool {
	testLimOnce.Do(func() {
		testLimInst = rate.NewLimiter(rate.Every(60_000_000_000/5), 5) // 5/menit, burst 5
	})
	return testLimInst.Allow()
}
