package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetChanStatus(t *testing.T) {
	assert.Equal(t, -1, getChanStatus(-1))
	assert.Equal(t, 0, getChanStatus(0))
	assert.Equal(t, 1, getChanStatus(5))
	assert.Equal(t, 2, getChanStatus(50))
}

func TestGetDiscussionStatus(t *testing.T) {
	assert.Equal(t, 0, getDiscussionStatus(0, 0))
	assert.Equal(t, 0, getDiscussionStatus(0, 1))
	assert.Equal(t, 0, getDiscussionStatus(1, 0))
	assert.Equal(t, 1, getDiscussionStatus(2, 0))
	assert.Equal(t, 2, getDiscussionStatus(0, 2))
	assert.Equal(t, 3, getDiscussionStatus(2, 2))
}
