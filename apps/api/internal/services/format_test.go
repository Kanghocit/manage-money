package services

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestCalcDeltaPct(t *testing.T) {
	assert.Equal(t, 0, calcDeltaPct(decimal.Zero, decimal.Zero))
	assert.Equal(t, 100, calcDeltaPct(decimal.Zero, decimal.NewFromInt(100)))
	prev := decimal.NewFromInt(100)
	curr := decimal.NewFromInt(120)
	assert.Equal(t, 20, calcDeltaPct(prev, curr))
}
