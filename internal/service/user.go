package service

import (
	"context"
	"errors"
	"fmt"
	"site-constructor/internal/apperrors"
	"site-constructor/internal/dto/user_context"
	"site-constructor/internal/models"
	"site-constructor/internal/repository"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	repo repository.User
}

func NewUserService(repo repository.User) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(input user_context.CreateUserInput) (*models.User, error) {
	existingUser, err := s.repo.GetByUsername(input.Username)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, apperrors.ErrUserExists
	}

	hashedPassword, err := hashPassword(input.Password)
	if err != nil {
		logrus.Errorf("[UserService] Failed to hash password: %v", err)
		return nil, err
	}

	createdUser, err := s.repo.Create(models.User{
		Name:     input.Username,
		Username: input.Username,
		Password: hashedPassword,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	logrus.Infof("[UserService] User created successfully: ID=%s, Username=%s, Name=%s", createdUser.ID, createdUser.Username, createdUser.Name)
	return createdUser, nil
}

func (s *UserService) GetAllUsers() ([]models.User, error) {
	users, err := s.repo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}
	return users, nil
}

func (s *UserService) GetByID(id uuid.UUID) (*models.User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return user, nil
}

func (s *UserService) GetUserByUsername(username string) (*models.User, error) {
	user, err := s.repo.GetByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return user, nil
}

func (s *UserService) Update(id uuid.UUID, input user_context.UpdateUserInput) (*models.User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	if input.Name != nil {
		user.Name = *input.Name
	}
	if input.Username != nil {
		user.Username = *input.Username
	}
	if input.Password != nil {
		if user.Password, err = hashPassword(*input.Password); err != nil {
			logrus.Errorf("[UserService] Failed to hash password: %v", err)
			return nil, err
		}
	}

	updatedUser, err := s.repo.Update(user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	logrus.Infof("[UserService] User updated successfully: ID=%s, Username=%s, Name=%s", user.ID, user.Username, user.Name)
	return updatedUser, nil
}

func (s *UserService) DeleteUser(id uuid.UUID, ctx context.Context) error {
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if err := s.repo.DeleteRefreshTokenByUserID(id, ctx); err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	logrus.Infof("[UserService] User deleted successfully by ID=%s", id)
	return nil
}

func hashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password is empty")
	}
	bs, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bs), nil
}
