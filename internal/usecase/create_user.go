package usecase

import (
	"context"
	"fmt"

	"github.com/vitao/geolocation-tracker/internal/domain/entity"
	"github.com/vitao/geolocation-tracker/internal/domain/repository"
	"github.com/vitao/geolocation-tracker/pkg/logger"
)

// CreateUserRequest representa a requisição para criar um usuário
type CreateUserRequest struct {
	ID      string `json:"id" binding:"required"`
	Name    string `json:"name" binding:"required"`
	Email   string `json:"email" binding:"required,email"`
	EventID string `json:"event_id" binding:"required"`
}

// CreateUserResponse representa a resposta da criação de usuário
type CreateUserResponse struct {
	UserID  string `json:"user_id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	EventID string `json:"event_id"`
	Message string `json:"message"`
}

// CreateUserUseCase representa o use case para criar usuários
type CreateUserUseCase struct {
	userRepo repository.UserRepository
	logger   logger.Logger
}

// NewCreateUserUseCase cria uma nova instância do use case
func NewCreateUserUseCase(
	userRepo repository.UserRepository,
	logger logger.Logger,
) *CreateUserUseCase {
	return &CreateUserUseCase{
		userRepo: userRepo,
		logger:   logger,
	}
}

// Execute executa o use case de criação de usuário
func (uc *CreateUserUseCase) Execute(ctx context.Context, req CreateUserRequest) (*CreateUserResponse, error) {
	// 1. Criar usuário
	user, err := entity.NewUser(req.ID, req.Name, req.Email)
	if err != nil {
		uc.logger.Error("Failed to create user entity", map[string]interface{}{
			"user_id": req.ID,
			"name":    req.Name,
			"email":   req.Email,
			"error":   err.Error(),
		})
		return nil, fmt.Errorf("invalid user data: %w", err)
	}

	// 2. Verificar se o usuário já existe
	existingUser, err := uc.userRepo.FindByID(ctx, user.ID())
	if err == nil && existingUser != nil {
		uc.logger.Info("User already exists", map[string]interface{}{
			"user_id": req.ID,
		})
		existingUserID := existingUser.ID()
		existingUserEmail := existingUser.Email()
		return &CreateUserResponse{
			UserID:  existingUserID.String(),
			Name:    existingUser.Name(),
			Email:   existingUserEmail.String(),
			EventID: req.EventID,
			Message: "User already exists",
		}, nil
	}

	// 3. Salvar usuário no repository
	if err := uc.userRepo.Save(ctx, user); err != nil {
		uc.logger.Error("Failed to save user", map[string]interface{}{
			"user_id": req.ID,
			"error":   err.Error(),
		})
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	uc.logger.Info("User created successfully", map[string]interface{}{
		"user_id": req.ID,
		"name":    req.Name,
		"email":   req.Email,
	})

	userID := user.ID()
	userEmail := user.Email()

	return &CreateUserResponse{
		UserID:  userID.String(),
		Name:    user.Name(),
		Email:   userEmail.String(),
		EventID: req.EventID,
		Message: "User created successfully",
	}, nil
}
