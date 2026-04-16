package service

import (
	"fmt"
	"time"

	"absensi-app/internal/model"
	"absensi-app/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo  *repository.UserRepository
	jwtSecret string
}

func NewAuthService(userRepo *repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

// Login authenticates user and returns JWT token
func (s *AuthService) Login(username, password string) (*model.LoginResponse, error) {
	// Find user
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, fmt.Errorf("user account is inactive")
	}

	// Check if account is locked
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		remainingTime := time.Until(*user.LockedUntil).Round(time.Minute)
		return nil, fmt.Errorf("account is locked due to too many failed login attempts. Try again in %v", remainingTime)
	}

	// If lock period has expired, reset the lock
	if user.LockedUntil != nil && time.Now().After(*user.LockedUntil) {
		if err := s.userRepo.ResetFailedLoginAttempts(user.ID); err != nil {
			// Log error but continue with login attempt
			fmt.Printf("Warning: failed to reset lock for user %d: %v\n", user.ID, err)
		}
		user.FailedLoginAttempts = 0
		user.LockedUntil = nil
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		// Increment failed attempts
		if incrementErr := s.userRepo.IncrementFailedLoginAttempts(user.ID); incrementErr != nil {
			fmt.Printf("Warning: failed to increment failed attempts for user %d: %v\n", user.ID, incrementErr)
		}

		// Check if we need to lock the account (5 failed attempts = 15 minutes lock)
		user.FailedLoginAttempts++
		if user.FailedLoginAttempts >= 5 {
			lockUntil := time.Now().Add(15 * time.Minute)
			lockUntilStr := lockUntil.Format("2006-01-02 15:04:05")
			if lockErr := s.userRepo.LockAccount(user.ID, lockUntilStr); lockErr != nil {
				fmt.Printf("Warning: failed to lock account for user %d: %v\n", user.ID, lockErr)
			}
			return nil, fmt.Errorf("too many failed login attempts. Account locked for 15 minutes")
		}

		remainingAttempts := 5 - user.FailedLoginAttempts
		return nil, fmt.Errorf("invalid credentials. %d attempts remaining before account lock", remainingAttempts)
	}

	// Successful login - reset failed attempts
	if user.FailedLoginAttempts > 0 {
		if err := s.userRepo.ResetFailedLoginAttempts(user.ID); err != nil {
			// Log error but continue with successful login
			fmt.Printf("Warning: failed to reset failed attempts for user %d: %v\n", user.ID, err)
		}
	}

	// Generate JWT token
	token, err := s.generateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &model.LoginResponse{
		Token: token,
		User:  *user,
	}, nil
}

// generateToken creates a JWT token for the user
func (s *AuthService) generateToken(user *model.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GetUserByID retrieves user by ID
func (s *AuthService) GetUserByID(userID int64) (*model.User, error) {
	return s.userRepo.FindByID(userID)
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
