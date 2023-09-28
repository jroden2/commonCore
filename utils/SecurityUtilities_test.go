package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRandSequence(t *testing.T) {
	var got1 = RandSequence(9)
	var got2 = RandSequence(9)

	assert.NotEqual(t, got1, got2)
	assert.Equal(t, 9, len(got1))
}

func TestGenerateID(t *testing.T) {
	var got1 = GenerateID()
	var got2 = GenerateID()

	assert.NotEqual(t, got1, got2)
	assert.Equal(t, 12, len(got1))
}
