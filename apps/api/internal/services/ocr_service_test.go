package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockOCRResult(t *testing.T) {
	r := mockOCRResult()
	assert.NotEmpty(t, r.Items)
	assert.NotEmpty(t, r.Items[0].Amount)
}
