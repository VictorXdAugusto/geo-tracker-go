package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/vitao/geolocation-tracker/internal/domain/entity"
	"github.com/vitao/geolocation-tracker/internal/usecase"
	"github.com/vitao/geolocation-tracker/internal/usecase/mocks"
)

// CreateUserUseCaseTestSuite define a suite de testes para CreateUserUseCase
type CreateUserUseCaseTestSuite struct {
	suite.Suite
	userRepo   *mocks.MockUserRepository
	logger     *mocks.MockLogger
	useCase    *usecase.CreateUserUseCase
	ctx        context.Context
	validUser  *entity.User
	validEmail entity.Email
	validID    entity.UserID
}

// SetupTest configura cada teste
func (suite *CreateUserUseCaseTestSuite) SetupTest() {
	suite.userRepo = new(mocks.MockUserRepository)
	suite.logger = new(mocks.MockLogger)
	suite.useCase = usecase.NewCreateUserUseCase(suite.userRepo, suite.logger)
	suite.ctx = context.Background()

	// Criar entidades válidas para reutilizar nos testes
	var err error
	suite.validUser, err = entity.NewUser("user123", "João Silva", "joao@example.com")
	suite.Require().NoError(err)

	validEmailPtr, err := entity.NewEmail("joao@example.com")
	suite.Require().NoError(err)
	suite.validEmail = *validEmailPtr

	validIDPtr, err := entity.NewUserID("user123")
	suite.Require().NoError(err)
	suite.validID = *validIDPtr
}

// TearDownTest limpa após cada teste
func (suite *CreateUserUseCaseTestSuite) TearDownTest() {
	suite.userRepo.AssertExpectations(suite.T())
	suite.logger.AssertExpectations(suite.T())
}

// TestCreateUser_Success testa criação bem-sucedida de usuário
func (suite *CreateUserUseCaseTestSuite) TestCreateUser_Success() {
	// Arrange
	request := usecase.CreateUserRequest{
		ID:      "user123",
		Name:    "João Silva",
		Email:   "joao@example.com",
		EventID: "event123",
	}

	// Mock: usuário não existe
	suite.userRepo.On("FindByID", mock.Anything, mock.AnythingOfType("entity.UserID")).
		Return(nil, errors.New("user not found"))

	// Mock: salvar usuário com sucesso
	suite.userRepo.On("Save", mock.Anything, mock.AnythingOfType("*entity.User")).
		Return(nil)

	// Mock: logs de sucesso
	suite.logger.On("Info", "User created successfully", mock.Anything).
		Return()

	// Act
	response, err := suite.useCase.Execute(suite.ctx, request)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response)
	assert.Equal(suite.T(), "user123", response.UserID)
	assert.Equal(suite.T(), "João Silva", response.Name)
	assert.Equal(suite.T(), "joao@example.com", response.Email)
	assert.Equal(suite.T(), "event123", response.EventID)
	assert.Equal(suite.T(), "User created successfully", response.Message)
}

// TestCreateUser_UserAlreadyExists testa quando usuário já existe
func (suite *CreateUserUseCaseTestSuite) TestCreateUser_UserAlreadyExists() {
	// Arrange
	request := usecase.CreateUserRequest{
		ID:      "user123",
		Name:    "João Silva",
		Email:   "joao@example.com",
		EventID: "event123",
	}

	// Mock: usuário já existe
	suite.userRepo.On("FindByID", mock.Anything, mock.AnythingOfType("entity.UserID")).
		Return(suite.validUser, nil)

	// Mock: log de usuário existente
	suite.logger.On("Info", "User already exists", mock.Anything).
		Return()

	// Act
	response, err := suite.useCase.Execute(suite.ctx, request)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response)
	assert.Equal(suite.T(), "user123", response.UserID)
	assert.Equal(suite.T(), "User already exists", response.Message)
}

// TestCreateUser_InvalidUserData testa com dados inválidos de usuário
func (suite *CreateUserUseCaseTestSuite) TestCreateUser_InvalidUserData() {
	testCases := []struct {
		name    string
		request usecase.CreateUserRequest
		wantErr string
	}{
		{
			name: "email inválido",
			request: usecase.CreateUserRequest{
				ID:      "user123",
				Name:    "João Silva",
				Email:   "email-invalido",
				EventID: "event123",
			},
			wantErr: "invalid user data",
		},
		{
			name: "ID vazio",
			request: usecase.CreateUserRequest{
				ID:      "",
				Name:    "João Silva",
				Email:   "joao@example.com",
				EventID: "event123",
			},
			wantErr: "invalid user data",
		},
		{
			name: "nome vazio",
			request: usecase.CreateUserRequest{
				ID:      "user123",
				Name:    "",
				Email:   "joao@example.com",
				EventID: "event123",
			},
			wantErr: "invalid user data",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Mock: log de erro esperado
			suite.logger.On("Error", "Failed to create user entity", mock.Anything).
				Return().Maybe() // Maybe() permite que não seja chamado se a validação falhar antes

			// Act
			response, err := suite.useCase.Execute(suite.ctx, tc.request)

			// Assert
			assert.Error(suite.T(), err)
			assert.Nil(suite.T(), response)
			assert.Contains(suite.T(), err.Error(), tc.wantErr)
		})
	}
}

// TestCreateUser_RepositorySaveError testa erro ao salvar no repositório
func (suite *CreateUserUseCaseTestSuite) TestCreateUser_RepositorySaveError() {
	// Arrange
	request := usecase.CreateUserRequest{
		ID:      "user123",
		Name:    "João Silva",
		Email:   "joao@example.com",
		EventID: "event123",
	}

	repositoryError := errors.New("database connection failed")

	// Mock: usuário não existe
	suite.userRepo.On("FindByID", mock.Anything, mock.AnythingOfType("entity.UserID")).
		Return(nil, errors.New("user not found"))

	// Mock: erro ao salvar
	suite.userRepo.On("Save", mock.Anything, mock.AnythingOfType("*entity.User")).
		Return(repositoryError)

	// Mock: log de erro
	suite.logger.On("Error", "Failed to save user", mock.Anything).
		Return()

	// Act
	response, err := suite.useCase.Execute(suite.ctx, request)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), response)
	assert.Contains(suite.T(), err.Error(), "failed to save user")
	assert.Contains(suite.T(), err.Error(), "database connection failed")
}

// TestNewCreateUserUseCase testa o construtor
func (suite *CreateUserUseCaseTestSuite) TestNewCreateUserUseCase() {
	// Act
	uc := usecase.NewCreateUserUseCase(suite.userRepo, suite.logger)

	// Assert
	assert.NotNil(suite.T(), uc)
}

// TestCreateUserUseCase executa toda a suite de testes
func TestCreateUserUseCase(t *testing.T) {
	suite.Run(t, new(CreateUserUseCaseTestSuite))
}
