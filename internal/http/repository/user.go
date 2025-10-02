package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pavelc4/auriya-todolist-go/internal/cache"
)

type User struct {
	ID             int64      `json:"id"`
	Email          string     `json:"email"`
	FullName       string     `json:"full_name"`
	Age            int        `json:"age"`
	Password       string     `json:"-"`
	AvatarURL      string     `json:"avatar_url"`
	Provider       string     `json:"provider"`
	ProviderUserID string     `json:"provider_user_id"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	LastLogin      *time.Time `json:"last_login,omitempty"`
}

type UserRepository struct {
	db    *pgxpool.Pool
	cache *cache.Service
}

func NewUserRepository(db *pgxpool.Pool, cache *cache.Service) *UserRepository {
	return &UserRepository{db: db, cache: cache}
}

func (r *UserRepository) GetByProviderUserID(ctx context.Context, provider, providerUserID string) (*User, error) {
	cacheKey := fmt.Sprintf("user:provider:%s:%s", provider, providerUserID)
	if cached, found := r.cache.Get(cacheKey); found {
		if user, ok := cached.(*User); ok {
			return user, nil
		}
	}

	const query = `
		SELECT id, email, full_name, avatar_url, provider, provider_user_id, password, age, created_at, updated_at, last_login
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
		&user.Password,
		&user.Age,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLogin,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	r.cache.Set(cacheKey, user, 5*time.Minute)
	r.cache.Set(fmt.Sprintf("user:email:%s", user.Email), user, 5*time.Minute)
	r.cache.Set(fmt.Sprintf("user:id:%d", user.ID), user, 5*time.Minute)

	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	cacheKey := fmt.Sprintf("user:email:%s", email)
	if cached, found := r.cache.Get(cacheKey); found {
		if user, ok := cached.(*User); ok {
			return user, nil
		}
	}

	const query = `
		SELECT id, email, full_name, avatar_url, provider, provider_user_id, password, age, created_at, updated_at, last_login
		FROM users
		WHERE email = $1
		LIMIT 1;
	`
	row := r.db.QueryRow(ctx, query, email)
	user := &User{}
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.FullName,
		&user.AvatarURL,
		&user.Provider,
		&user.ProviderUserID,
		&user.Password,
		&user.Age,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLogin,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	r.cache.Set(cacheKey, user, 5*time.Minute)
	r.cache.Set(fmt.Sprintf("user:provider:%s:%s", user.Provider, user.ProviderUserID), user, 5*time.Minute)
	r.cache.Set(fmt.Sprintf("user:id:%d", user.ID), user, 5*time.Minute)

	return user, nil
}

func (r *UserRepository) Create(ctx context.Context, user *User) error {
	const query = `
		INSERT INTO users (email, full_name, avatar_url, provider, provider_user_id, password, age, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, query,
		user.Email,
		user.FullName,
		user.AvatarURL,
		user.Provider,
		user.ProviderUserID,
		user.Password,
		user.Age,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err == nil {
		r.cache.Set(fmt.Sprintf("user:id:%d", user.ID), user, 5*time.Minute)
		r.cache.Set(fmt.Sprintf("user:email:%s", user.Email), user, 5*time.Minute)
		if user.Provider != "" && user.ProviderUserID != "" {
			r.cache.Set(fmt.Sprintf("user:provider:%s:%s", user.Provider, user.ProviderUserID), user, 5*time.Minute)
		}
	}

	return err
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID int64) error {
	const query = `
		UPDATE users SET last_login = NOW(), updated_at = NOW() WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, userID)
	if err == nil {
		// Invalidate cache to reflect the new last_login time on next fetch
		r.cache.Delete(fmt.Sprintf("user:id:%d", userID))
	}
	return err
}
