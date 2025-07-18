package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommander(t *testing.T) {
	// arrange
	c := NewCommander()
	c.AddCommand("test", func(map[string]any) any {
		return "hi"
	})

	// act && assert
	assert.Len(t, c.Commands(), 1)
	assert.NotNil(t, c.Command("test"))
	assert.Nil(t, c.Command("booyeah"))
}
