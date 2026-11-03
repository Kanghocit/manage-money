package services

import (
	"testing"

	"github.com/manage-money/api/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestPrimaryRole(t *testing.T) {
	user := models.User{Roles: []models.Role{{Name: "user"}}}
	assert.Equal(t, "user", primaryRole(user))
	user.Roles = append(user.Roles, models.Role{Name: "admin"})
	assert.Equal(t, "admin", primaryRole(user))
}
