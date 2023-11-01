package monitor

import (
	"context"
	"github.com/stretchr/testify/assert"
	. "github.com/stretchr/testify/mock"
	"testing"
)

func TestMockExample(t *testing.T) {
	t.Run("test mock example", func(t *testing.T) {
		singleFlight := NewMockSingleFlight(t)
		singleFlight.On("BeginRecord", Anything, Anything).Return(nil)

		rec := singleFlight.BeginRecord(context.Background(), "test")
		assert.Nil(t, rec)

	})
}
