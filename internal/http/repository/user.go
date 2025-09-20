package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID             int64
	Email          string
	FullName       string
	AvatarURL      string
	Provider       string
	ProviderUserID string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	LastLogin      time.Time
}

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {

	return &UserRepository{db: db}
}

func (r *UserRepository) GetByProviderUserID(ctx context.Context, provider, providerUserID string) (*User, error) {
	const query = `
		SELECT id, email, full_name, avatar_url, provider, provider_user_id, created_at, updated_at, last_login
		FROM users
		WHERE provider = $1 AND provider_user_id = $2
		LIMIT 1;
	`

	row := r.db.QueryRow(ctx, query, provider, providerUserID)

	user := &User{}
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.FullName,
		&user.AvatarURL,
		&user.Provider,
		&user.ProviderUserID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLogin,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // user Not Found
		}
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) Create(ctx context.Context, user *User) error {
	const query = `
    INSERT INTO users (email, full_name, avatar_url, provider, provider_user_id, created_at, updated_at)
    VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
    RETURNING id, created_at, updated_at
    `
	err := r.db.QueryRow(ctx, query,
		user.Email,
		user.FullName,
		user.AvatarURL,
		user.Provider,
		user.ProviderUserID).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	return err
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID int64) error {
	const query = `
    UPDATE users SET last_login = NOW(), updated_at = NOW() WHERE id = $1
    `
	_, err := r.db.Exec(ctx, query, userID)
	return err
}
