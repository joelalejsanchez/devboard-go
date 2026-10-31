package postgres

import (
	"context"
	"devboard/internal/domain"
	generated "devboard/internal/repository/postgres/generated"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository implements UseCase.UserRepository
type UserRepository struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// NewUserRepository constructor
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		pool:    pool,
		queries: generated.New(pool),
	}
}

func toDomainUser(u generated.User) *domain.User {
	return &domain.User{
		ID:           u.ID.String(),
		Name:         u.Name,
		Email:        u.Email,
		PasswordHash: u.Email,
		CreatedAt:    u.CreatedAt.Time,
		UpdatedAt:    u.UpdatedAt.Time,
	}
}

// Create new user in database
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	result, err := r.queries.CreateUser(ctx, generated.CreateUserParams{
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		Name:         user.Name,
	})

	if err != nil {
		return fmt.Errorf("creando usuario en BD: %w", err)
	}

	user = toDomainUser(result)
	return nil
}

// GetByID get one user by id
func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var pgID pgtype.UUID
	if err := pgID.Scan(id); err != nil {
		return nil, fmt.Errorf("id invalido %s: %w", id, domain.ErrInvalidInput)
	}

	result, err := r.queries.GetUserByID(ctx, pgID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("buscando usuario en BD: %w", domain.ErrNotFound)

		}
		return nil, fmt.Errorf("buscando usuario en BD: %w", err)
	}

	return toDomainUser(result), nil
}

// GetByEmail get one user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	result, err := r.queries.GetUserByEmail(ctx, email)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("buscando usuario en BD: %w", domain.ErrNotFound)

		}
		return nil, fmt.Errorf("buscando usuario en BD: %w", err)
	}

	return toDomainUser(result), nil
}
