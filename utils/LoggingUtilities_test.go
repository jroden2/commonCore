package utils

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInitialiseLogger(t *testing.T) {
	t.Run("Should return a zerolog.logger", func(t *testing.T) {
		logger := InitialiseLogger()
		assert.IsType(t, zerolog.Logger{}, logger)
	})
}

func TestLogError(t *testing.T) {
	t.Run("should log the correct error", func(t *testing.T) {
		var buf bytes.Buffer
		logger := InitialiseLogger(&buf)
		LogError("TestFunc", errors.New("this is a test error"), logger)
		got := buf.String()
		fmt.Println(got)
		assert.NotEqual(t, "", got)
	})
}

func TestLogInfo(t *testing.T) {
	t.Run("should log the correct message", func(t *testing.T) {
		var buf bytes.Buffer
		logger := InitialiseLogger(&buf)
		LogInfo("TestFunc", "this is a test message", logger)
		got := buf.String()
		fmt.Println(got)
		assert.NotEqual(t, "", got)
	})
}
