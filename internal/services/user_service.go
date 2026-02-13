package services

import (
	"context"
	"errors"

	"github.com/pipeline-arch/app/internal/models"
	"github.com/pipeline-arch/app/internal/repository"
	"github.com/pipeline-arch/app/pkg/errors"
	"github.com/pipeline-arch/app/pkg/metrics"
	"github.com/rs/zerolog"
)

// UserService handles user business logic
type UserService struct {
	repo    repository.UserRepository
	log     *zerolog.Logger
	metrics *metrics.Metrics
}

// NewUserService creates a new user service
func NewUserService(repo repository.UserRepository, log *zerolog.Logger, m *metrics.Metrics) *UserService {
	return &UserService{
		repo:    repo,
		log:     log,
		metrics: m,
	}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, req *models.UserCreateRequest) (*models.UserResponse, error) {
	s.log.Info().Str("email", req.Email).Msg("Creating new user")

	// Check if user with email already exists
	existing, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		s.log.Error().Err(err).Str("email", req.Email).Msg("Error checking existing user")
		return nil, errors.ErrInternalServer
	}
	if existing != nil {
		return nil, errors.NewAppError(
			errors.ErrCodeConflict,
			"User with this email already exists",
			req.Email,
			nil,
		)
	}

	// Create new user
	user := models.NewUser(req.Email, req.Name, req.Role)
	if err := s.repo.Create(ctx, user); err != nil {
		s.log.Error().Err(err).Str("email", req.Email).Msg("Error creating user")
		return nil, errors.ErrInternalServer
	}

	// Update metrics
	s.metrics.IncUsers()
	s.metrics.IncOperation("create", "success")

	s.log.Info().Str("user_id", user.ID).Str("email", user.Email).Msg("User created successfully")

	return user.ToResponse(), nil
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(ctx context.Context, id string) (*models.UserResponse, error) {
	s.log.Info().Str("user_id", id).Msg("Getting user")

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.Error().Err(err).Str("user_id", id).Msg("Error getting user")
		return nil, errors.ErrInternalServer
	}
	if user == nil {
		return nil, errors.NotFoundError("User", id)
	}

	s.metrics.IncOperation("get", "success")
	return user.ToResponse(), nil
}

// ListUsers retrieves a paginated list of users
func (s *UserService) ListUsers(ctx context.Context, page, pageSize int) (*models.UserListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	s.log.Info().Int("page", page).Int("page_size", pageSize).Msg("Listing users")

	offset := (page - 1) * pageSize
	users, err := s.repo.List(ctx, pageSize, offset)
	if err != nil {
		s.log.Error().Err(err).Msg("Error listing users")
		return nil, errors.ErrInternalServer
	}

	total, err := s.repo.Count(ctx)
	if err != nil {
		s.log.Error().Err(err).Msg("Error counting users")
		return nil, errors.ErrInternalServer
	}

	totalPages := (total + pageSize - 1) / pageSize

	userResponses := make([]*models.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = user.ToResponse()
	}

	s.metrics.IncOperation("list", "success")

	return &models.UserListResponse{
		Users:      userResponses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateUser updates an existing user
func (s *UserService) UpdateUser(ctx context.Context, id string, req *models.UserUpdateRequest) (*models.UserResponse, error) {
	s.log.Info().Str("user_id", id).Msg("Updating user")

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.Error().Err(err).Str("user_id", id).Msg("Error getting user for update")
		return nil, errors.ErrInternalServer
	}
	if user == nil {
		return nil, errors.NotFoundError("User", id)
	}

	// Update fields
	if req.Email != nil {
		// Check if email is already taken by another user
		existing, err := s.repo.GetByEmail(ctx, *req.Email)
		if err != nil {
			s.log.Error().Err(err).Str("email", *req.Email).Msg("Error checking email")
			return nil, errors.ErrInternalServer
		}
		if existing != nil && existing.ID != id {
			return nil, errors.NewAppError(
				errors.ErrCodeConflict,
				"Email is already in use",
				*req.Email,
				nil,
			)
		}
		user.Email = *req.Email
	}
	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Role != nil {
		user.Role = *req.Role
	}
	if req.Active != nil {
		user.Active = *req.Active
	}

	if err := s.repo.Update(ctx, user); err != nil {
		s.log.Error().Err(err).Str("user_id", id).Msg("Error updating user")
		return nil, errors.ErrInternalServer
	}

	s.metrics.IncOperation("update", "success")
	s.log.Info().Str("user_id", id).Msg("User updated successfully")

	return user.ToResponse(), nil
}

// DeleteUser deletes a user
func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	s.log.Info().Str("user_id", id).Msg("Deleting user")

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.Error().Err(err).Str("user_id", id).Msg("Error getting user for delete")
		return errors.ErrInternalServer
	}
	if user == nil {
		return errors.NotFoundError("User", id)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.log.Error().Err(err).Str("user_id", id).Msg("Error deleting user")
		return errors.ErrInternalServer
	}

	s.metrics.IncOperation("delete", "success")
	s.log.Info().Str("user_id", id).Msg("User deleted successfully")

	return nil
}

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidInput     = errors.New("invalid input")
)