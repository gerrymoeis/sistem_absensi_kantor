package service

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"

	"github.com/Kagami/go-face"
	"absensi-app/internal/model"
	"absensi-app/internal/repository"
)

type FaceService struct {
	recognizer  *face.Recognizer
	repo        *repository.FaceRepository
	userRepo    *repository.UserRepository
	threshold   float64 // Distance threshold for matching (default: 0.6)
	environment string  // development or production
}

// NewFaceService creates a new face service
func NewFaceService(modelsPath string, repo *repository.FaceRepository, userRepo *repository.UserRepository, environment string) (*FaceService, error) {
	rec, err := face.NewRecognizer(modelsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize recognizer: %w", err)
	}

	return &FaceService{
		recognizer:  rec,
		repo:        repo,
		userRepo:    userRepo,
		threshold:   0.6, // Validated optimal threshold from POC
		environment: environment,
	}, nil
}

// Close closes the face recognizer
func (s *FaceService) Close() {
	if s.recognizer != nil {
		s.recognizer.Close()
	}
}

// EnrollFace enrolls a new face for a user
func (s *FaceService) EnrollFace(userID int64, imageData string) error {
	// Decode base64 image
	imgBytes, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// Decode image to check quality
	img, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return fmt.Errorf("failed to decode image format: %w", err)
	}

	// Quality checks
	if err := s.checkImageQuality(img); err != nil {
		return fmt.Errorf("image quality check failed: %w", err)
	}

	// Save image temporarily for go-face processing
	tmpFile, err := os.CreateTemp("", "face-*.jpg")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.Write(imgBytes); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close()

	// Recognize face using file
	faces, err := s.recognizer.RecognizeFile(tmpFile.Name())
	if err != nil {
		return fmt.Errorf("failed to recognize face: %w", err)
	}

	if len(faces) == 0 {
		return fmt.Errorf("no face detected in image")
	}

	if len(faces) > 1 {
		return fmt.Errorf("multiple faces detected, please ensure only one face in image")
	}

	// Get face descriptor (128-dimensional encoding)
	descriptor := faces[0].Descriptor

	// Serialize descriptor to bytes
	encodingBytes, err := s.serializeDescriptor(descriptor)
	if err != nil {
		return fmt.Errorf("failed to serialize descriptor: %w", err)
	}

	// Calculate quality score (based on face rectangle size)
	qualityScore := s.calculateQualityScore(faces[0])

	// Save encoding
	encoding := &model.FaceEncoding{
		UserID:       userID,
		Encoding:     encodingBytes,
		QualityScore: qualityScore,
	}

	if err := s.repo.SaveEncoding(encoding); err != nil {
		return fmt.Errorf("failed to save encoding: %w", err)
	}

	return nil
}

// RecognizeFace recognizes a face and returns the matched user
func (s *FaceService) RecognizeFace(imageData string, ipAddress string, livenessPassed bool) (*model.FaceRecognitionResponse, error) {
	// Decode base64 image
	imgBytes, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Calculate image hash (for replay attack prevention)
	imageHash := s.calculateImageHash(imgBytes)

	// Check if image hash exists (only in production)
	// In development, skip this check to allow repeated testing
	if s.environment == "production" {
		exists, err := s.repo.CheckImageHashExists(imageHash, 24*60)
		if err != nil {
			return nil, fmt.Errorf("failed to check image hash: %w", err)
		}
		if exists {
			return &model.FaceRecognitionResponse{
				Matched:    false,
				Confidence: 0,
				Message:    "Image already used recently (possible replay attack)",
			}, nil
		}
	}

	// Decode image to check quality
	_, _, err = image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image format: %w", err)
	}

	// Save image temporarily for go-face processing
	tmpFile, err := os.CreateTemp("", "face-*.jpg")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.Write(imgBytes); err != nil {
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close()

	// Recognize face using file
	faces, err := s.recognizer.RecognizeFile(tmpFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to recognize face: %w", err)
	}

	if len(faces) == 0 {
		// Log failed attempt
		s.logAttempt(nil, false, 0, livenessPassed, imageHash, ipAddress)
		return &model.FaceRecognitionResponse{
			Matched:    false,
			Confidence: 0,
			Message:    "No face detected in image",
		}, nil
	}

	if len(faces) > 1 {
		// Log failed attempt
		s.logAttempt(nil, false, 0, livenessPassed, imageHash, ipAddress)
		return &model.FaceRecognitionResponse{
			Matched:    false,
			Confidence: 0,
			Message:    "Multiple faces detected, please ensure only one face in frame",
		}, nil
	}

	// Get face descriptor
	descriptor := faces[0].Descriptor

	// Get all stored encodings
	allEncodings, err := s.repo.GetAllEncodings()
	if err != nil {
		return nil, fmt.Errorf("failed to get encodings: %w", err)
	}

	if len(allEncodings) == 0 {
		// Log failed attempt
		s.logAttempt(nil, false, 0, livenessPassed, imageHash, ipAddress)
		return &model.FaceRecognitionResponse{
			Matched:    false,
			Confidence: 0,
			Message:    "No enrolled faces in system",
		}, nil
	}

	// Find best match
	var bestMatch *model.FaceEncoding
	var bestDistance float64 = math.MaxFloat64

	for i := range allEncodings {
		storedDescriptor, err := s.deserializeDescriptor(allEncodings[i].Encoding)
		if err != nil {
			continue // Skip invalid encodings
		}

		distance := s.euclideanDistance(descriptor, storedDescriptor)
		if distance < bestDistance {
			bestDistance = distance
			bestMatch = &allEncodings[i]
		}
	}

	// Check if best match is below threshold
	if bestDistance < s.threshold {
		// Match found!
		user, err := s.userRepo.FindByID(bestMatch.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}

		// Check if user is active
		if !user.IsActive {
			// Log failed attempt
			s.logAttempt(&bestMatch.UserID, false, bestDistance, livenessPassed, imageHash, ipAddress)
			return &model.FaceRecognitionResponse{
				Matched:    false,
				Confidence: 1.0 - bestDistance, // Convert distance to confidence
				Message:    "User account is inactive",
			}, nil
		}

		// Log successful attempt
		s.logAttempt(&bestMatch.UserID, true, bestDistance, livenessPassed, imageHash, ipAddress)

		return &model.FaceRecognitionResponse{
			Matched:    true,
			UserID:     &user.ID,
			FullName:   user.FullName,
			Confidence: 1.0 - bestDistance, // Convert distance to confidence (0-1 scale)
			Message:    "Face recognized successfully",
		}, nil
	}

	// No match found
	s.logAttempt(nil, false, bestDistance, livenessPassed, imageHash, ipAddress)
	return &model.FaceRecognitionResponse{
		Matched:    false,
		Confidence: 1.0 - bestDistance,
		Message:    "Face not recognized",
	}, nil
}

// GetUserEncodings retrieves all encodings for a user
func (s *FaceService) GetUserEncodings(userID int64) ([]model.FaceEncoding, error) {
	return s.repo.GetUserEncodings(userID)
}

// DeleteUserFaceData deletes all face data for a user
func (s *FaceService) DeleteUserFaceData(userID int64) error {
	return s.repo.DeleteUserEncodings(userID)
}

// GetEncodingStats retrieves encoding statistics (admin)
func (s *FaceService) GetEncodingStats() ([]model.FaceEncodingInfo, error) {
	return s.repo.GetEncodingStats()
}

// Helper functions

func (s *FaceService) checkImageQuality(img image.Image) error {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Check minimum resolution (lowered for LFW dataset compatibility)
	// LFW images are typically 94x125, which is sufficient for face recognition
	if width < 80 || height < 80 {
		return fmt.Errorf("image resolution too low (minimum 80x80)")
	}

	// Check maximum resolution (prevent memory issues)
	if width > 4096 || height > 4096 {
		return fmt.Errorf("image resolution too high (maximum 4096x4096)")
	}

	return nil
}

func (s *FaceService) calculateQualityScore(f face.Face) float64 {
	// Quality score based on face rectangle size
	// Larger face = better quality
	rect := f.Rectangle
	faceArea := rect.Dx() * rect.Dy()

	// Normalize to 0-100 scale
	// Assume optimal face size is 200x200 = 40000 pixels
	score := float64(faceArea) / 40000.0 * 100.0

	// Cap at 100
	if score > 100 {
		score = 100
	}

	return score
}

func (s *FaceService) serializeDescriptor(desc face.Descriptor) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, desc)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *FaceService) deserializeDescriptor(data []byte) (face.Descriptor, error) {
	var desc face.Descriptor
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, &desc)
	if err != nil {
		return desc, err
	}
	return desc, nil
}

func (s *FaceService) euclideanDistance(a, b face.Descriptor) float64 {
	var sum float32
	for i := range a {
		diff := a[i] - b[i]
		sum += diff * diff
	}
	return math.Sqrt(float64(sum))
}

func (s *FaceService) calculateImageHash(imgBytes []byte) string {
	hash := sha256.Sum256(imgBytes)
	return fmt.Sprintf("%x", hash)
}

func (s *FaceService) logAttempt(userID *int64, matched bool, distance float64, livenessPassed bool, imageHash string, ipAddress string) {
	attempt := &model.FaceAttempt{
		UserID:         userID,
		Matched:        matched,
		Confidence:     1.0 - distance, // Convert distance to confidence
		LivenessPassed: livenessPassed,
		ImageHash:      imageHash,
		IPAddress:      ipAddress,
	}

	// Log attempt (ignore errors, logging should not block main flow)
	_ = s.repo.LogAttempt(attempt)
}
