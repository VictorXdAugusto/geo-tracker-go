package entity

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/vitao/geolocation-tracker/internal/domain/valueobject"
)

// User representa um usuário do sistema de geolocalização
// Entidade = tem identidade única (ID), pode mudar estado, tem ciclo de vida
// Agregado Root = responsável por manter consistência das suas partes
type User struct {
	id        UserID                 // Identidade única
	name      string                 // Nome do usuário
	email     Email                  // Email (value object)
	createdAt *valueobject.Timestamp // Quando foi criado
	updatedAt *valueobject.Timestamp // Última atualização
}

// UserID representa o identificador único do usuário
type UserID struct {
	value string
}

// Email representa um email válido
type Email struct {
	value string
}

// Constantes de validação
const (
	MinNameLength = 2
	MaxNameLength = 100
)

// Regex para validação de email
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// Erros específicos do domínio User
var (
	ErrEmptyUserID    = errors.New("user ID cannot be empty")
	ErrInvalidEmail   = errors.New("invalid email format")
	ErrInvalidName    = errors.New("invalid name")
	ErrNameTooShort   = errors.New("name too short")
	ErrNameTooLong    = errors.New("name too long")
	ErrUserIDNotFound = errors.New("user ID not found")
)

// NewUserID cria um novo UserID
func NewUserID(id string) (*UserID, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrEmptyUserID
	}

	return &UserID{value: strings.TrimSpace(id)}, nil
}

// Value retorna o valor do UserID
func (uid *UserID) Value() string {
	return uid.value
}

// String implementa fmt.Stringer
func (uid *UserID) String() string {
	return uid.value
}

// Equals compara dois UserIDs
func (uid *UserID) Equals(other *UserID) bool {
	if other == nil {
		return false
	}
	return uid.value == other.value
}

// NewEmail cria um novo Email válido
func NewEmail(email string) (*Email, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	if !emailRegex.MatchString(email) {
		return nil, fmt.Errorf("%w: %s", ErrInvalidEmail, email)
	}

	return &Email{value: email}, nil
}

// Value retorna o valor do Email
func (e *Email) Value() string {
	return e.value
}

// String implementa fmt.Stringer
func (e *Email) String() string {
	return e.value
}

// Equals compara dois Emails
func (e *Email) Equals(other *Email) bool {
	if other == nil {
		return false
	}
	return e.value == other.value
}

// NewUser cria um novo usuário (Factory Method)
// Garante que o usuário é criado em estado válido
func NewUser(id, name, email string) (*User, error) {
	// Validar e criar UserID
	userID, err := NewUserID(id)
	if err != nil {
		return nil, err
	}

	// Validar e criar Email
	userEmail, err := NewEmail(email)
	if err != nil {
		return nil, err
	}

	// Validar nome
	if err := validateName(name); err != nil {
		return nil, err
	}

	now := valueobject.Now()

	return &User{
		id:        *userID,
		name:      strings.TrimSpace(name),
		email:     *userEmail,
		createdAt: now,
		updatedAt: now,
	}, nil
}

// validateName valida o nome do usuário
func validateName(name string) error {
	name = strings.TrimSpace(name)

	if name == "" {
		return ErrInvalidName
	}

	if len(name) < MinNameLength {
		return fmt.Errorf("%w: minimum %d characters", ErrNameTooShort, MinNameLength)
	}

	if len(name) > MaxNameLength {
		return fmt.Errorf("%w: maximum %d characters", ErrNameTooLong, MaxNameLength)
	}

	return nil
}

// Getters (expõem estado de forma segura)
func (u *User) ID() UserID {
	return u.id
}

func (u *User) Name() string {
	return u.name
}

func (u *User) Email() Email {
	return u.email
}

func (u *User) CreatedAt() *valueobject.Timestamp {
	return u.createdAt
}

func (u *User) UpdatedAt() *valueobject.Timestamp {
	return u.updatedAt
}

// UpdateName atualiza o nome do usuário (comportamento da entidade)
func (u *User) UpdateName(newName string) error {
	if err := validateName(newName); err != nil {
		return err
	}

	// Só atualizar se realmente mudou
	trimmedName := strings.TrimSpace(newName)
	if u.name != trimmedName {
		u.name = trimmedName
		u.updatedAt = valueobject.Now()
	}

	return nil
}

// UpdateEmail atualiza o email do usuário
func (u *User) UpdateEmail(newEmail string) error {
	email, err := NewEmail(newEmail)
	if err != nil {
		return err
	}

	// Só atualizar se realmente mudou
	if !u.email.Equals(email) {
		u.email = *email
		u.updatedAt = valueobject.Now()
	}

	return nil
}

// String implementa fmt.Stringer
func (u *User) String() string {
	return fmt.Sprintf("User{ID: %s, Name: %s, Email: %s}",
		u.id.Value(), u.name, u.email.Value())
}

// Equals compara dois usuários pela identidade (ID)
func (u *User) Equals(other *User) bool {
	if other == nil {
		return false
	}
	return u.id.Equals(&other.id)
}
