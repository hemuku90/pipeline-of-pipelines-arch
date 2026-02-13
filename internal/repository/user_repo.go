package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/pipeline-arch/app/internal/models"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*models.User, error)
	Count(ctx context.Context) (int, error)
	Close() error
}

// PostgresUserRepository implements UserRepository for PostgreSQL
type PostgresUserRepository struct {
	db        *sql.DB
	tableName string
}

// NewPostgresUserRepository creates a new PostgreSQL user repository
func NewPostgresUserRepository(db *sql.DB, tableName string) *PostgresUserRepository {
	return &PostgresUserRepository{
		db:        db,
		tableName: tableName,
	}
}

// Create creates a new user
func (r *PostgresUserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, email, name, role, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Name,
		user.Role,
		user.Active,
		user.CreatedAt,
		user.UpdatedAt,
	)
	return err
}

// GetByID retrieves a user by ID
func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, email, name, role, active, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Role,
		&user.Active,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetByEmail retrieves a user by email
func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, name, role, active, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Role,
		&user.Active,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Update updates an existing user
func (r *PostgresUserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET email = $1, name = $2, role = $3, active = $4, updated_at = $5
		WHERE id = $6
	`
	user.UpdatedAt = time.Now().UTC()
	_, err := r.db.ExecContext(ctx, query,
		user.Email,
		user.Name,
		user.Role,
		user.Active,
		user.UpdatedAt,
		user.ID,
	)
	return err
}

// Delete deletes a user by ID
func (r *PostgresUserRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// List retrieves a list of users
func (r *PostgresUserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	query := `
		SELECT id, email, name, role, active, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Name,
			&user.Role,
			&user.Active,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

// Count returns the total number of users
func (r *PostgresUserRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM users`
	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

// Close closes the database connection
func (r *PostgresUserRepository) Close() error {
	return r.db.Close()
}

// InMemoryUserRepository provides a simple in-memory implementation for testing
type InMemoryUserRepository struct {
	users map[string]*models.User
}

// NewInMemoryUserRepository creates a new in-memory user repository
func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users: make(map[string]*models.User),
	}
}

// Create creates a new user
func (r *InMemoryUserRepository) Create(ctx context.Context, user *models.User) error {
	r.users[user.ID] = user
	return nil
}

// GetByID retrieves a user by ID
func (r *InMemoryUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	user, exists := r.users[id]
	if !exists {
		return nil, nil
	}
	return user, nil
}

// GetByEmail retrieves a user by email
func (r *InMemoryUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, nil
}

// Update updates an existing user
func (r *InMemoryUserRepository) Update(ctx context.Context, user *models.User) error {
	r.users[user.ID] = user
	return nil
}

// Delete deletes a user by ID
func (r *InMemoryUserRepository) Delete(ctx context.Context, id string) error {
	delete(r.users, id)
	return nil
}

// List retrieves a list of users
func (r *InMemoryUserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	users := make([]*models.User, 0, limit)
	count := 0
	for _, user := range r.users {
		if count >= offset {
			users = append(users, user)
			if len(users) >= limit {
				break
			}
		}
		count++
	}
	return users, nil
}

// Count returns the total number of users
func (r *InMemoryUserRepository) Count(ctx context.Context) (int, error) {
	return len(r.users), nil
}

// Close is a no-op for in-memory repository
func (r *InMemoryUserRepository) Close() error {
	return nil
}