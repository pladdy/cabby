package cabby

import (
	"context"

	"github.com/gofrs/uuid"
)

// Key type for context ids; per context docuentation, use a Key type for context Keys
type Key int

const (
	// KeyBytes stores the bytes written for a response
	KeyBytes Key = iota

	// KeyTransactionID to track a request from http request to response (including service calls)
	KeyTransactionID

	// KeyUser for storing a User struct
	KeyUser
)

// TakeBytes returns the bytes stored in a context
func TakeBytes(ctx context.Context) int {
	bytes, ok := ctx.Value(KeyBytes).(int)
	if !ok {
		return 0
	}
	return bytes
}

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

// WithBytes decorates a context with bytes written
func WithBytes(ctx context.Context, bytes int) context.Context {
	return context.WithValue(ctx, KeyBytes, bytes)
}

// WithTransactionID decorates a context with a transaction id
func WithTransactionID(ctx context.Context, transactionID uuid.UUID) context.Context {
	return context.WithValue(ctx, KeyTransactionID, transactionID)
}

// WithUser decorates a context with values from the user struct
func WithUser(ctx context.Context, u User) context.Context {
	return context.WithValue(ctx, KeyUser, u)
}
