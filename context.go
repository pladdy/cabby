package cabby

import (
	"context"

	"github.com/gofrs/uuid"
)

// Key type for context ids; per context docuentation, use a Key type for context Keys
type Key int

const (
	// KeyTransactionID to track a request from http request to response (including service calls)
	KeyTransactionID Key = 0

	// KeyUser for storing a User struct
	KeyUser Key = 1
)

// TakeTransactionID returns the transaction id stored in a context
func TakeTransactionID(ctx context.Context) uuid.UUID {
	id, ok := ctx.Value(KeyTransactionID).(uuid.UUID)
	if !ok {
		return uuid.UUID{}
	}
	return id
}

// TakeUser returns the user stored in a context
func TakeUser(ctx context.Context) User {
	u, ok := ctx.Value(KeyUser).(User)
	if !ok {
		return User{}
	}
	return u
}

// WithTransactionID decorates a context with a transaction id
func WithTransactionID(ctx context.Context, transactionID uuid.UUID) context.Context {
	ctx = context.WithValue(ctx, KeyTransactionID, transactionID)
	return ctx
}

// WithUser decorates a context with values from the user struct
func WithUser(ctx context.Context, u User) context.Context {
	ctx = context.WithValue(ctx, KeyUser, u)
	return ctx
}
