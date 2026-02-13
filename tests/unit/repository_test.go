package unit

import (
	"context"
	"testing"

	"github.com/pipeline-arch/app/internal/models"
	"github.com/pipeline-arch/app/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryUserRepository(t *testing.T) {
	repo := repository.NewInMemoryUserRepository()
	ctx := context.Background()

	t.Run("CreateUser", func(t *testing.T) {
		user := models.NewUser("test@example.com", "Test User", "admin")
		err := repo.Create(ctx, user)
		require.NoError(t, err)

		// Verify user was created
		retrieved, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.Equal(t, user.Email, retrieved.Email)
		assert.Equal(t, user.Name, retrieved.Name)
		assert.Equal(t, user.Role, retrieved.Role)
	})

	t.Run("GetByEmail", func(t *testing.T) {
		repo := repository.NewInMemoryUserRepository()
		user := models.NewUser("email@test.com", "Email User", "user")
		err := repo.Create(ctx, user)
		require.NoError(t, err)

		retrieved, err := repo.GetByEmail(ctx, "email@test.com")
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.Equal(t, "email@test.com", retrieved.Email)
	})

	t.Run("GetByIDNotFound", func(t *testing.T) {
		repo := repository.NewInMemoryUserRepository()
		retrieved, err := repo.GetByID(ctx, "non-existent-id")
		require.NoError(t, err)
		assert.Nil(t, retrieved)
	})

	t.Run("UpdateUser", func(t *testing.T) {
		repo := repository.NewInMemoryUserRepository()
		user := models.NewUser("update@test.com", "Original Name", "user")
		err := repo.Create(ctx, user)
		require.NoError(t, err)

		user.Name = "Updated Name"
		user.Role = "admin"
		err = repo.Update(ctx, user)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.Equal(t, "Updated Name", retrieved.Name)
		assert.Equal(t, "admin", retrieved.Role)
	})

	t.Run("DeleteUser", func(t *testing.T) {
		repo := repository.NewInMemoryUserRepository()
		user := models.NewUser("delete@test.com", "Delete User", "user")
		err := repo.Create(ctx, user)
		require.NoError(t, err)

		err = repo.Delete(ctx, user.ID)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Nil(t, retrieved)
	})

	t.Run("ListUsers", func(t *testing.T) {
		repo := repository.NewInMemoryUserRepository()

		// Create multiple users
		for i := 0; i < 5; i++ {
			user := models.NewUser(
				"user"+string(rune('0'+i))+"@test.com",
				"User "+string(rune('0'+i)),
				"user",
			)
			err := repo.Create(ctx, user)
			require.NoError(t, err)
		}

		users, err := repo.List(ctx, 10, 0)
		require.NoError(t, err)
		assert.Len(t, users, 5)
	})

	t.Run("CountUsers", func(t *testing.T) {
		repo := repository.NewInMemoryUserRepository()

		count, err := repo.Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, count)

		user := models.NewUser("count@test.com", "Count User", "user")
		err = repo.Create(ctx, user)
		require.NoError(t, err)

		count, err = repo.Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("Close", func(t *testing.T) {
		repo := repository.NewInMemoryUserRepository()
		err := repo.Close()
		assert.NoError(t, err)
	})
}

func TestModels(t *testing.T) {
	t.Run("NewUser", func(t *testing.T) {
		user := models.NewUser("new@test.com", "New User", "admin")
		assert.NotEmpty(t, user.ID)
		assert.Equal(t, "new@test.com", user.Email)
		assert.Equal(t, "New User", user.Name)
		assert.Equal(t, "admin", user.Role)
		assert.True(t, user.Active)
		assert.False(t, user.CreatedAt.IsZero())
		assert.False(t, user.UpdatedAt.IsZero())
	})

	t.Run("UserToResponse", func(t *testing.T) {
		user := models.NewUser("response@test.com", "Response User", "viewer")
		response := user.ToResponse()
		assert.Equal(t, user.ID, response.ID)
		assert.Equal(t, user.Email, response.Email)
		assert.Equal(t, user.Name, response.Name)
		assert.Equal(t, user.Role, response.Role)
		assert.Equal(t, user.Active, response.Active)
	})

	t.Run("UserCreateRequest", func(t *testing.T) {
		req := models.UserCreateRequest{
			Email: "create@test.com",
			Name:  "Create User",
			Role:  "user",
		}
		assert.Equal(t, "create@test.com", req.Email)
		assert.Equal(t, "Create User", req.Name)
		assert.Equal(t, "user", req.Role)
	})

	t.Run("UserUpdateRequest", func(t *testing.T) {
		name := "Updated Name"
		active := false
		req := models.UserUpdateRequest{
			Name:   &name,
			Active: &active,
		}
		assert.Equal(t, "Updated Name", *req.Name)
		assert.Equal(t, false, *req.Active)
	})
}