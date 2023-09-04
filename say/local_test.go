package say

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSay(t *testing.T) {
	err := say(context.Background(), "Hello, World!")
	require.NoError(t, err)
}
