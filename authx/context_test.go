package authx

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithIdentityAndCurrentIdentity(t *testing.T) {
	base := context.Background()
	identity := NewIdentity("u-1", "user", "Alice")

	ctx := WithIdentity(base, identity)
	got, ok := CurrentIdentity(ctx)
	assert.True(t, ok)
	assert.Equal(t, identity.ID(), got.ID())
}

func TestRequireIdentity(t *testing.T) {
	_, err := RequireIdentity(context.Background())
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrNoIdentity))

	ctx := WithIdentity(context.Background(), NewIdentity("u-1", "user", "Alice"))
	got, err := RequireIdentity(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "u-1", got.ID())
}

func TestWithIdentityNilIdentity(t *testing.T) {
	base := context.Background()
	ctx := WithIdentity(base, nil)

	_, ok := CurrentIdentity(ctx)
	assert.False(t, ok)
}
