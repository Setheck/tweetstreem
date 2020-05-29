package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHistory(t *testing.T) {
	history := NewHistory()
	assert.NotNil(t, history)
	assert.Zero(t, history.LastIdx())
	last, ok := history.Last(1)
	assert.Nil(t, last)
	assert.False(t, ok)

	obj1 := struct{}{}
	obj2 := struct{}{}
	history.Log(obj1)
	assert.Equal(t, 1, history.LastIdx())
	history.Log(obj2)
	assert.Equal(t, 2, history.LastIdx())

	if obj, ok := history.Last(1); true {
		assert.Equal(t, obj1, obj)
		assert.True(t, ok)
	}

	if obj, ok := history.Last(2); true {
		assert.Equal(t, obj2, obj)
		assert.True(t, ok)
	}

	if obj, ok := history.Last(3); true {
		assert.Nil(t, obj)
		assert.False(t, ok)
	}

	history.Clear()

	if obj, ok := history.Last(1); true {
		assert.Nil(t, obj)
		assert.False(t, ok)
	}

	if obj, ok := history.Last(2); true {
		assert.Nil(t, obj)
		assert.False(t, ok)
	}

	assert.Zero(t, history.LastIdx())
}
