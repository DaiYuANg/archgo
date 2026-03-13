package authx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrincipalContext(t *testing.T) {
	principal := Principal{ID: "u1"}
	ctx := WithPrincipal(context.Background(), principal)

	got, ok := PrincipalFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, principal, got)

	typed, ok := PrincipalFromContextAs[Principal](ctx)
	assert.True(t, ok)
	assert.Equal(t, principal, typed)
}
