package service

import (
	"fmt"
	"regexp"

	"absensi-app/internal/model"
	"absensi-app/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// CreateUser creates a new user with validation
func (s *UserService) CreateUser(username, password, fullName, role string, isActive bool) (*model.User, error) {
	// Validate username
	if err := s.validateUsername(username); err != nil {
		return nil, err
	}

	// Check if username exists
	exists, err := s.userRepo.CheckUsernameExists(username, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to check username: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("username already exists")
	}

	// Validate password
	if err := s.validatePassword(password); err != nil {
		return nil, err
	}

	// Validate full name
	if fullName == "" {
		return nil, fmt.Errorf("full name is required")
	}

	// Validate role
	if role != "admin" && role != "employee" {
		return nil, fmt.Errorf("role must be 'admin' or 'employee'")
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &model.User{
		Username:     username,
		PasswordHash: string(passwordHash),
		FullName:     fullName,
		Role:         role,
		IsActive:     isActive,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// UpdateUser updates user information
func (s *UserService) UpdateUser(userID int64, fullName, role string, isActive bool) (*model.User, error) {
	// Get existing user
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Validate full name
	if fullName == "" {
		return nil, fmt.Errorf("full name is required")
	}

	// Validate role
	if role != "admin" && role != "employee" {
		return nil, fmt.Errorf("role must be 'admin' or 'employee'")
	}

	// Update user
	user.FullName = fullName
	user.Role = role
	user.IsActive = isActive

	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// ResetPassword resets user password
func (s *UserService) ResetPassword(userID int64, newPassword string) error {
	// Validate password
	if err := s.validatePassword(newPassword); err != nil {
		return err
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	if err := s.userRepo.UpdatePassword(userID, string(passwordHash)); err != nil {
		return fmt.Errorf("failed to reset password: %w", err)
	}

	return nil
}

// DeleteUser soft deletes a user
func (s *UserService) DeleteUser(userID int64) error {
	// Check if user exists
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Soft delete
	if err := s.userRepo.Delete(userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// GetUser gets user by ID
func (s *UserService) GetUser(userID int64) (*model.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return user, nil
}

// validateUsername validates username format
func (s *UserService) validateUsername(username string) error {
	if len(username) < 3 || len(username) > 50 {
		return fmt.Errorf("username must be between 3 and 50 characters")
	}

	// Alphanumeric and underscore only
	matched, _ := regexp.MatchString("^[a-zA-Z0-9_]+$", username)
	if !matched {
		return fmt.Errorf("username must contain only alphanumeric characters and underscore")
	}

	return nil
}

// validatePassword validates password strength with complexity rules
func (s *UserService) validatePassword(password string) error {
	// Minimum length
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	// Maximum length (prevent DoS via bcrypt)
	if len(password) > 72 {
		return fmt.Errorf("password must not exceed 72 characters")
	}

	// Check for at least one uppercase letter
	hasUpper, _ := regexp.MatchString("[A-Z]", password)
	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	// Check for at least one lowercase letter
	hasLower, _ := regexp.MatchString("[a-z]", password)
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	// Check for at least one digit
	hasDigit, _ := regexp.MatchString("[0-9]", password)
	if !hasDigit {
		return fmt.Errorf("password must contain at least one number")
	}

	// Check for at least one special character
	hasSpecial, _ := regexp.MatchString("[!@#$%^&*()_+\\-=\\[\\]{};':\"\\\\|,.<>/?]", password)
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character (!@#$%^&*()_+-=[]{}...)")
	}

	return nil
}
