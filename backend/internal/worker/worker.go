package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"inventariskantor/internal/config"
	"inventariskantor/internal/telegram"
)

const (
	DefaultQueue = "queue:inventaris:tasks"

	TaskTypeTelegramMovement = "telegram:movement"
	TaskTypeSecurityEvent    = "security:event"
)

// Task represents an asynchronous background job
type Task struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	CreatedAt time.Time       `json:"created_at"`
	Retry     int             `json:"retry"`
}

// TelegramMovementPayload holds arguments for movement notifications
type TelegramMovementPayload struct {
	TxType   string `json:"tx_type"`
	SKU      string `json:"sku"`
	Name     string `json:"name"`
	Quantity int32  `json:"quantity"`
	Unit     string `json:"unit"`
	Prev     int32  `json:"prev"`
	Next     int32  `json:"next"`
	Person   string `json:"person"`
	Notes    string `json:"notes"`
}

// SecurityEventPayload holds arguments for async security event logging
type SecurityEventPayload struct {
	EventType    string    `json:"event_type"`
	IP           string    `json:"ip"`
	Endpoint     string    `json:"endpoint"`
	Method       string    `json:"method"`
	AttemptCount int       `json:"attempt_count"`
	Status       string    `json:"status"`
	UserAgent    string    `json:"user_agent"`
	Details      string    `json:"details"`
	CreatedAt    time.Time `json:"created_at"`
}

// NewRedisClient creates and verifies a Redis connection
func NewRedisClient(cfg *config.Config) (*redis.Client, error) {
	var opts *redis.Options
	if cfg.RedisURL != "" {
		var err error
		opts, err = redis.ParseURL(cfg.RedisURL)
		if err != nil {
			return nil, fmt.Errorf("invalid REDIS_URL: %w", err)
		}
	} else {
		opts = &redis.Options{
			Addr:         net.JoinHostPort(cfg.RedisHost, cfg.RedisPort),
			Password:     cfg.RedisPassword,
			DB:           cfg.RedisDB,
			DialTimeout:  3 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
		}
	}

	client := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, err
	}

	return client, nil
}

// Queue handles task dispatching to Redis
type Queue struct {
	rdb *redis.Client
}

func NewQueue(rdb *redis.Client) *Queue {
	return &Queue{rdb: rdb}
}

// Enqueue adds a task to the default queue. Falls back silently if Redis is offline.
func (q *Queue) Enqueue(ctx context.Context, taskType string, payload any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload failed: %w", err)
	}

	task := Task{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Type:      taskType,
		Payload:   raw,
		CreatedAt: time.Now(),
	}

	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("marshal task failed: %w", err)
	}

	if q.rdb != nil {
		return q.rdb.LPush(ctx, DefaultQueue, data).Err()
	}
	return nil
}

// Worker processes background tasks from Redis
type Worker struct {
	rdb      *redis.Client
	pool     *pgxpool.Pool
	tgBot    *telegram.Bot
	handlers map[string]func(ctx context.Context, payload []byte) error
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

// NewWorker creates a worker instance
func NewWorker(rdb *redis.Client, pool *pgxpool.Pool, tgBot *telegram.Bot) *Worker {
	w := &Worker{
		rdb:      rdb,
		pool:     pool,
		tgBot:    tgBot,
		handlers: make(map[string]func(ctx context.Context, payload []byte) error),
		stopCh:   make(chan struct{}),
	}

	w.registerDefaultHandlers()
	return w
}

func (w *Worker) registerDefaultHandlers() {
	// Handle Telegram movement notifications
	w.handlers[TaskTypeTelegramMovement] = func(ctx context.Context, payload []byte) error {
		if w.tgBot == nil {
			return nil
		}
		var p TelegramMovementPayload
		if err := json.Unmarshal(payload, &p); err != nil {
			return err
		}
		w.tgBot.NotifyMovement(ctx, p.TxType, p.SKU, p.Name, p.Quantity, p.Unit, p.Prev, p.Next, p.Person, p.Notes)
		return nil
	}

	// Handle Security Event persisting to PostgreSQL
	w.handlers[TaskTypeSecurityEvent] = func(ctx context.Context, payload []byte) error {
		if w.pool == nil {
			return nil
		}
		var p SecurityEventPayload
		if err := json.Unmarshal(payload, &p); err != nil {
			return err
		}
		_, err := w.pool.Exec(ctx,
			`INSERT INTO security_events (event_type, ip, endpoint, method, attempt_count, status, user_agent, details, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			p.EventType, p.IP, p.Endpoint, p.Method, p.AttemptCount, p.Status, p.UserAgent, p.Details, p.CreatedAt)
		return err
	}
}

// Start spawns background consumer goroutines
func (w *Worker) Start(ctx context.Context, concurrency int) {
	if w.rdb == nil {
		slog.Warn("worker skipped: redis not connected")
		return
	}
	if concurrency <= 0 {
		concurrency = 3
	}

	slog.Info("starting background workers", "concurrency", concurrency)
	for i := 0; i < concurrency; i++ {
		w.wg.Add(1)
		go w.runConsumer(ctx, i+1)
	}

	// Start periodic scheduled checks (e.g. low stock, security log cleanup)
	w.wg.Add(1)
	go w.runScheduler(ctx)
}

func (w *Worker) runConsumer(ctx context.Context, id int) {
	defer w.wg.Done()
	slog.Debug("worker consumer started", "worker_id", id)

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		default:
			// Pop task with timeout
			res, err := w.rdb.BRPop(ctx, 2*time.Second, DefaultQueue).Result()
			if err != nil {
				if err == redis.Nil || ctx.Err() != nil {
					continue
				}
				time.Sleep(500 * time.Millisecond)
				continue
			}

			if len(res) < 2 {
				continue
			}

			var task Task
			if err := json.Unmarshal([]byte(res[1]), &task); err != nil {
				slog.Error("worker invalid task payload", "err", err)
				continue
			}

			handler, exists := w.handlers[task.Type]
			if !exists {
				slog.Warn("worker unhandled task type", "type", task.Type)
				continue
			}

			taskCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
			if err := handler(taskCtx, task.Payload); err != nil {
				slog.Error("worker task failed", "type", task.Type, "err", err)
			} else {
				slog.Debug("worker task completed", "type", task.Type, "id", task.ID)
			}
			cancel()
		}
	}
}

// runScheduler executes periodic background maintenance
func (w *Worker) runScheduler(ctx context.Context) {
	defer w.wg.Done()
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		case <-ticker.C:
			// Cleanup old security events (> 30 days) to keep DB lean
			if w.pool != nil {
				cleanCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
				_, err := w.pool.Exec(cleanCtx, `DELETE FROM security_events WHERE created_at < NOW() - INTERVAL '30 days'`)
				if err != nil {
					slog.Warn("failed to cleanup old security events", "err", err)
				}
				cancel()
			}
		}
	}
}

// Stop gracefully shuts down workers
func (w *Worker) Stop() {
	close(w.stopCh)
	w.wg.Wait()
	slog.Info("background workers stopped gracefully")
}
