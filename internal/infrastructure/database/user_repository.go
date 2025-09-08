package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/vitao/geolocation-tracker/internal/domain/entity"
	"github.com/vitao/geolocation-tracker/internal/domain/repository"
	"github.com/vitao/geolocation-tracker/pkg/logger"
)

// userRepository implementa repository.UserRepository usando PostgreSQL
type userRepository struct {
	db     *DB
	logger logger.Logger
}

// NewUserRepository cria uma nova instância do repository de usuários
func NewUserRepository(db *DB, logger logger.Logger) repository.UserRepository {
	return &userRepository{
		db:     db,
		logger: logger,
	}
}

// Save persiste um usuário (INSERT ou UPDATE)
func (r *userRepository) Save(ctx context.Context, user *entity.User) error {
	// Query para UPSERT (INSERT ON CONFLICT UPDATE)
	query := `
		INSERT INTO users (id, name, email, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			email = EXCLUDED.email,
			updated_at = EXCLUDED.updated_at
	`

	// Extrair valores para evitar problemas com métodos
	userID := user.ID()
	userEmail := user.Email()

	_, err := r.db.Connection().ExecContext(ctx, query,
		userID.Value(),
		user.Name(),
		userEmail.Value(),
		user.CreatedAt().Time(),
		user.UpdatedAt().Time(),
	)

	if err != nil {
		r.logger.Error("Failed to save user",
			"user_id", userID.Value(),
			"error", err,
		)
		return fmt.Errorf("failed to save user %s: %w", userID.Value(), err)
	}

	r.logger.Debug("User saved successfully",
		"user_id", userID.Value(),
		"name", user.Name(),
	)

	return nil
}

// FindByID busca usuário por ID
func (r *userRepository) FindByID(ctx context.Context, id entity.UserID) (*entity.User, error) {
	query := `
		SELECT id, name, email, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var userID, name, email string
	var createdAt, updatedAt sql.NullTime

	err := r.db.Connection().QueryRowContext(ctx, query, id.Value()).Scan(
		&userID, &name, &email, &createdAt, &updatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %s", id.Value())
		}
		r.logger.Error("Failed to find user by ID",
			"user_id", id.Value(),
			"error", err,
		)
		return nil, fmt.Errorf("failed to find user %s: %w", id.Value(), err)
	}

	// Reconstruir entidade User
	user, err := r.scanToUser(userID, name, email, createdAt, updatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to reconstruct user %s: %w", id.Value(), err)
	}

	return user, nil
}

// FindByEmail busca usuário por email
func (r *userRepository) FindByEmail(ctx context.Context, email entity.Email) (*entity.User, error) {
	query := `
		SELECT id, name, email, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var userID, name, emailStr string
	var createdAt, updatedAt sql.NullTime

	err := r.db.Connection().QueryRowContext(ctx, query, email.Value()).Scan(
		&userID, &name, &emailStr, &createdAt, &updatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found with email: %s", email.Value())
		}
		r.logger.Error("Failed to find user by email",
			"email", email.Value(),
			"error", err,
		)
		return nil, fmt.Errorf("failed to find user by email %s: %w", email.Value(), err)
	}

	// Reconstruir entidade User
	user, err := r.scanToUser(userID, name, emailStr, createdAt, updatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to reconstruct user with email %s: %w", email.Value(), err)
	}

	return user, nil
}

// Exists verifica se usuário existe
func (r *userRepository) Exists(ctx context.Context, id entity.UserID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`

	var exists bool
	err := r.db.Connection().QueryRowContext(ctx, query, id.Value()).Scan(&exists)
	if err != nil {
		r.logger.Error("Failed to check user existence",
			"user_id", id.Value(),
			"error", err,
		)
		return false, fmt.Errorf("failed to check if user %s exists: %w", id.Value(), err)
	}

	return exists, nil
}

// Delete remove usuário
func (r *userRepository) Delete(ctx context.Context, id entity.UserID) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.Connection().ExecContext(ctx, query, id.Value())
	if err != nil {
		r.logger.Error("Failed to delete user",
			"user_id", id.Value(),
			"error", err,
		)
		return fmt.Errorf("failed to delete user %s: %w", id.Value(), err)
	}

	// Verificar se alguma linha foi afetada
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found: %s", id.Value())
	}

	r.logger.Info("User deleted successfully",
		"user_id", id.Value(),
	)

	return nil
}

// FindAll retorna todos os usuários com paginação
func (r *userRepository) FindAll(ctx context.Context, limit, offset int) ([]*entity.User, error) {
	query := `
		SELECT id, name, email, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Connection().QueryContext(ctx, query, limit, offset)
	if err != nil {
		r.logger.Error("Failed to find all users",
			"limit", limit,
			"offset", offset,
			"error", err,
		)
		return nil, fmt.Errorf("failed to find users: %w", err)
	}
	defer rows.Close()

	users := make([]*entity.User, 0)

	for rows.Next() {
		var userID, name, email string
		var createdAt, updatedAt sql.NullTime

		if err := rows.Scan(&userID, &name, &email, &createdAt, &updatedAt); err != nil {
			r.logger.Error("Failed to scan user row",
				"error", err,
			)
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		user, err := r.scanToUser(userID, name, email, createdAt, updatedAt)
		if err != nil {
			r.logger.Error("Failed to reconstruct user from row",
				"user_id", userID,
				"error", err,
			)
			continue // Pular usuários inválidos
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	r.logger.Debug("Found users",
		"count", len(users),
		"limit", limit,
		"offset", offset,
	)

	return users, nil
}

// scanToUser converte dados do banco para entidade User
func (r *userRepository) scanToUser(userID, name, email string, _, _ sql.NullTime) (*entity.User, error) {
	// Esta é uma função de reconstrução - precisamos usar um factory interno
	// Por enquanto, vamos usar o factory público (idealmente teríamos um método interno)
	user, err := entity.NewUser(userID, name, email)
	if err != nil {
		return nil, err
	}

	// NOTA: Em uma implementação mais sofisticada, teríamos métodos para
	// reconstruir a entidade com timestamps originais do banco
	// Por agora, os timestamps serão recriados

	return user, nil
}
