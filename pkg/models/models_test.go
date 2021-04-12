package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRWItem(t *testing.T) {
	assert := assert.New(t)
	tmpModel := NewRWItem()
	assert.NotNil(tmpModel)

}

func TestCloseChan(t *testing.T) {
	// assert := assert.New(t)
	// tmpModel := NewRWItem()
	// for
}
