package handlers

import (
	"bytes"
	"context"
	"io"

	"github.com/alexedwards/scs/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"inventariskantor/internal/service"
)

type Handlers struct {
	pool            *pgxpool.Pool
	svc             *service.Service
	sess            *scs.SessionManager
	broker          *Broker
	driveLastRun    string
	driveLastResult string
}

func New(pool *pgxpool.Pool, sess *scs.SessionManager) *Handlers {
	return &Handlers{pool: pool, svc: service.NewService(pool), sess: sess, broker: NewBroker()}
}

var (
	_ = context.Background
	_ = io.EOF
	_ = bytes.MinRead
)
