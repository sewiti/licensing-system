package license

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIdentifier(t *testing.T) {
	_, err := Identifier()
	assert.NoError(t, err)
}
