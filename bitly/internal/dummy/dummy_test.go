package dummy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdd(t *testing.T) {
	assert.Equal(t, 3, Add(1, 2))
	require.Equal(t, 3, Add(1, 2))
}
