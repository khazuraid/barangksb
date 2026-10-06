package auth

import "context"

type ctxKey string

// UserKey is the context key storing the authenticated *User.
const UserKey ctxKey = "user"

func WithUser(ctx context.Context, u *User) context.Context {
	return context.WithValue(ctx, UserKey, u)
}
