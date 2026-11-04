package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockOCRResult(t *testing.T) {
	r := mockOCRResult()
	assert.Empty(t, r.Items)
}
