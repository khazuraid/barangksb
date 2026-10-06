package handlers

import (
	"github.com/alexedwards/scs/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handlers struct {
	pool            *pgxpool.Pool
	sess            *scs.SessionManager
	broker          *Broker
	driveLastRun    string
	driveLastResult string
}

func New(pool *pgxpool.Pool, sess *scs.SessionManager) *Handlers {
	return &Handlers{pool: pool, sess: sess, broker: NewBroker()}
}
