package cabby

import (
	"context"
	"testing"

	"github.com/gofrs/uuid"
)

func TestContextTransactionID(t *testing.T) {
	ctx := context.Background()

	transactionID := uuid.Must(uuid.NewV4())
	ctx = WithTransactionID(ctx, transactionID)
	result := TakeTransactionID(ctx)

	if result != transactionID {
		t.Error("Got:", result, "Expected:", transactionID)
	}
}

func TestContextUser(t *testing.T) {
	ctx := context.Background()
	u := User{Email: "foo"}

	ctx = WithUser(ctx, u)
	result := TakeUser(ctx)

	if result.Email != u.Email {
		t.Error("Got:", result.Email, "Expected:", u.Email)
	}
}
