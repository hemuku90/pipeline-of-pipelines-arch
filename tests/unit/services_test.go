package unit

import (
	"context"
	"testing"

	"github.com/pipeline-arch/app/internal/models"
	"github.com/pipeline-arch/app/internal/repository"
	"github.com/pipeline-arch/app/internal/services"
	"github.com/pipeline-arch/app/pkg/errors"
	"github.com/pipeline-arch/app/pkg/logger"
	"github.com/pipeline-arch/app/pkg/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockMetrics is a mock implementation of metrics
type MockMetrics struct {
	mock.Mock
}

func (m *MockMetrics) IncRequest(path string) {
	m.Called(path)
}

func TestUserService(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewInMemoryUserRepository()
	log := logger.New("debug")
	_ = metrics.New("pipeline-arch-test", 9091)
	svc := services.NewUserService(repo, log, nil)

	t.Run("CreateUser", func(t *testing.T) {
		repo := repository.NewInMemoryUserRepository()
		svc := services.NewUserService(repo, log, nil)

		req := &models.UserCreateRequest{
			Email: "test@example.com",
			Name:  "Test User",
			Role:  "admin",
		}

		user, err := svc.CreateUser(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, "Test User", user.Name)
		assert.Equal(t, "admin", user.Role)
	})

	t.Run("CreateUserAlreadyExists", func(t *testing.T) {
		repo := repository.NewInMemoryUserRepository()
		svc := services.NewUserService(repo, log, nil)

		req := &models.UserCreateRequest{
			Email: "duplicate@example.com",
			Name:  "Test User",
			Role:  "user",
		}

		// Create first user
		_, err := svc.CreateUser(ctx, req)
		require.NoError(t, err)

		// Try to create duplicate
		_, err = svc.CreateUser(ctx, req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("GetUser", func(t *testing.T) {
		repo := repository.NewInMemoryUserRepository()
		svc := services.NewUserService(repo, log, nil)

		// Create a user first
		createReq := &models.UserCreateRequest{
			Email: "get@example.com",
			Name:  "Get User",
			Role:  "user",
		}
		created, err := svc.CreateUser(ctx, createReq)
		require.NoError(t, err)

		// Get the user
		user, err := svc.GetUser(ctx, created.ID)
		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, created.ID, user.ID)
		assert.Equal(t, "get@example.com", user.Email)
	})

	t.Run("GetUserNotFound", func(t *testing.T) {
		repo := repository.NewInMemoryUserRepository()
		svc := services.NewUserService(repo, log, nil)

		user, err := svc.GetUser(ctx, "non-existent-id")
		require.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("ListUsers", func(t *testing.T) {
		repo := repository.NewInMemoryUserRepository()
		svc := services.NewUserService(repo, log, nil)

		// Create multiple users
		for i := 0; i < 5; i++ {
			req := &models.UserCreateRequest{
				Email: "list" + string(rune('0'+i)) + "@example.com",
				Name:  "User " + string(rune('0'+i)),
				Role:  "user",
			}
			_, err := svc.CreateUser(ctx, req)
			require.NoError(t, err)
		}

		result, err := svc.ListUsers(ctx, 1, 10)
		require.NoError(t, err)
		assert.Len(t, result.Users, 5)
		assert.Equal(t, 5, result.Total)
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 10, result.PageSize)
	})

	t.Run("UpdateUser", func(t *testing.T) {
		repo := repository.NewInMemoryUserRepository()
		svc := services.NewUserService(repo, log, nil)

		// Create a user
		created, err := svc.CreateUser(ctx, &models.UserCreateRequest{
			Email: "update@example.com",
			Name:  "Original Name",
			Role:  "user",
		})
		require.NoError(t, err)

		// Update the user
		newName := "Updated Name"
		newRole := "admin"
		updateReq := &models.UserUpdateRequest{
			Name: &newName,
			Role: &newRole,
		}

		updated, err := svc.UpdateUser(ctx, created.ID, updateReq)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", updated.Name)
		assert.Equal(t, "admin", updated.Role)
	})

	t.Run("UpdateUserNotFound", func(t *testing.T) {
		repo := repository.NewInMemoryUserRepository()
		svc := services.NewUserService(repo, log, nil)

		name := "Updated Name"
		updateReq := &models.UserUpdateRequest{
			Name: &name,
		}

		err := svc.UpdateUser(ctx, "non-existent-id", updateReq)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("DeleteUser", func(t *testing.T) {
		repo := repository.NewInMemoryUserRepository()
		svc := services.NewUserService(repo, log, nil)

		// Create a user
		created, err := svc.CreateUser(ctx, &models.UserCreateRequest{
			Email: "delete@example.com",
			Name:  "Delete User",
			Role:  "user",
		})
		require.NoError(t, err)

		// Delete the user
		err = svc.DeleteUser(ctx, created.ID)
		require.NoError(t, err)

		// Verify user is deleted
		user, err := svc.GetUser(ctx, created.ID)
		require.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("DeleteUserNotFound", func(t *testing.T) {
		repo := repository.NewInMemoryUserRepository()
		svc := services.NewUserService(repo, log, nil)

		err := svc.DeleteUser(ctx, "non-existent-id")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestAppErrors(t *testing.T) {
	t.Run("NotFoundError", func(t *testing.T) {
		err := errors.NotFoundError("TestResource", "123")
		assert.Contains(t, err.Error(), "not found")
		assert.Contains(t, err.Error(), "123")
	})

	t.Run("HTTPStatus", func(t *testing.T) {
		err := errors.NotFoundError("Test", "id")
		assert.Equal(t, 404, errors.HTTPStatus(err))

		err = errors.NewAppError(400, "Bad Request", "detail", nil)
		assert.Equal(t, 400, errors.HTTPStatus(err))
	})

	t.Run("Is", func(t *testing.T) {
		err := errors.NotFoundError("Test", "id")
		assert.True(t, errors.Is(err, errors.ErrNotFound))
		assert.False(t, errors.Is(err, errors.ErrUnauthorized))
	})

	t.Run("AppError", func(t *testing.T) {
		err := errors.NewAppError(500, "Internal Error", "detail", nil)
		assert.Equal(t, 500, err.Code)
		assert.Equal(t, "Internal Error", err.Message)
		assert.Equal(t, "detail", err.Detail)
	})
}